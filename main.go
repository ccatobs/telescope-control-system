package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

const (
	// poll the ACU status at 1 Hz
	statusUpdateDuration = 1000 * time.Millisecond

	// max time waiting to queue command
	commandBusyTimeout = 500 * time.Millisecond

	// http connection timeout
	connectionTimeout = 1000 * time.Millisecond
)

func init() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)
}

func logError(err error) {
	if err != nil {
		slog.Error(err.Error())
	}
}

func logNewCmd(id uint64, desc string) {
	if len(desc) > 200 {
		desc = fmt.Sprintf("%.200s...", desc)
	}
	slog.Info("got command", "cmd", id, "desc", desc)
}

func logCmdError(id uint64, err error) {
	if err != nil {
		slog.Error(err.Error(), "cmd", id)
	}
}

// Command IDs come from a simple incrementing counter. It's seeded with
// the startup time in milliseconds, so IDs stay unique across restarts.
var lastCmdID atomic.Uint64

func init() {
	lastCmdID.Store(uint64(time.Now().UnixMilli()))
}

func newCmdID() uint64 {
	return lastCmdID.Add(1)
}

// queuedCmd is a Command tagged with its ID.
type queuedCmd struct {
	id  uint64
	cmd Command
}

// jsonResponse writes a status response. A nonzero id is included in the response.
func jsonResponse(w http.ResponseWriter, id uint64, err error, statusCode int) {
	var response struct {
		S  string `json:"status"`
		ID uint64 `json:"id,omitempty"`
		M  string `json:"message,omitempty"`
	}

	response.ID = id
	if err != nil {
		response.S = "error"
		response.M = err.Error()
	} else {
		response.S = "ok"
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logError(err)
	}
}

func getenv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func main() {
	acuHost := getenv("FYST_ACU_HOST", "172.16.5.194")
	acuPort := getenv("FYST_ACU_PORT", "8100")
	acuMonitorPort := getenv("FYST_ACU_MONITOR_PORT", "8110")
	acuAdminPort := getenv("FYST_ACU_ADMIN_PORT", "8080")
	apiAddr := getenv("FYST_TCS_ADDR", ":5600")

	acu := NewACU(acuHost, acuPort, acuMonitorPort, acuAdminPort)
	tel := NewTelescope(acu)

	// report immediately any ACU problems
	err := tel.UpdateStatus()
	if err != nil {
		logError(err)
	}
	err = tel.CalibrateTime()
	if err != nil {
		logError(err)
	}
	err = tel.Ready()
	if err != nil {
		logError(err)
	}

	type MeasurementFloat struct {
		Name        string
		Description string
		Unit        string
		Value       float64
		Created     time.Time
	}

	var tel_pos = []MeasurementFloat{
		{
			Name:        "Elevation",
			Description: "Telescope height above sea level",
			Unit:        "meters",
			Value:       FYST_ELEVATION_METERS,
			Created:     time.Now(),
		},
		{
			Name:        "Latitude",
			Description: "Telescope latitude",
			Unit:        "degrees",
			Value:       FYST_LATITUDE_DEG,
			Created:     time.Now(),
		},
		{
			Name:        "Longitude",
			Description: "Telescope longitude with positive east",
			Unit:        "degrees",
			Value:       FYST_LONGITUDE_EAST_DEG,
			Created:     time.Now(),
		},
	}
	// XXX:DEBUG fake pointing model
	tel.pointing.azOffset = 0
	tel.pointing.elOffset = 0

	// command queue
	cmds := make(chan queuedCmd)

	// abort signal
	abort := make(chan chan bool)

	// main loop
	go func() {
		for {
			// wait for command
			var qc queuedCmd
		waitForCmdLoop:
			for {
				select {
				case qc = <-cmds:
					break waitForCmdLoop
				case <-time.After(statusUpdateDuration):
					err := tel.UpdateStatus()
					if err != nil {
						logError(err)
					}
				case c := <-abort:
					log.Print("ignoring abort")
					c <- false
				}
			}

			cmd := qc.cmd
			logNewCmd(qc.id, fmt.Sprintf("%#v", cmd))

			if err := tel.Ready(); err != nil {
				logCmdError(qc.id, err)
				continue
			}

			// start command
			ctx, cancel := context.WithCancel(context.Background())
			isDone, err := cmd.Start(ctx, tel)
			if err != nil {
				logCmdError(qc.id, err)
				cancel()
				continue
			}

			// wait for command to finish
			for done := false; !done; {
				select {
				case <-time.After(statusUpdateDuration):
					err = tel.UpdateStatus()
					if err != nil {
						break // select statement
					}
					done, err = isDone(tel)
				case c := <-abort:
					slog.Info("aborting command", "cmd", qc.id)
					c <- true
					done = true
					cancel()
					err = tel.Stop()
				}
				if err != nil {
					logCmdError(qc.id, err)
					break
				}
			}

			slog.Info("command done", "cmd", qc.id)
		}
	}()

	// build http API
	mux := http.NewServeMux()

	mux.Handle("GET /healthz", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := tel.Ready()
		statusCode := http.StatusServiceUnavailable
		if err == nil {
			statusCode = http.StatusOK
		}
		jsonResponse(w, 0, err, statusCode)
	}))

	mux.HandleFunc("/abort", func(w http.ResponseWriter, req *http.Request) {
		var err error
		var statusCode int

		var id uint64
		if req.Method == "POST" {
			id = newCmdID()
			logNewCmd(id, "abort")
			c := make(chan bool)
			abort <- c
			if <-c {
				statusCode = http.StatusOK
			} else {
				err = fmt.Errorf("nothing to abort")
				statusCode = http.StatusConflict // not sure if this is the most appropriate code
			}
		} else {
			err = fmt.Errorf("method not POST")
			statusCode = http.StatusMethodNotAllowed
		}

		jsonResponse(w, id, err, statusCode)
	})

	mux.HandleFunc("/acu/status", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			err := fmt.Errorf("method not GET")
			jsonResponse(w, 0, err, http.StatusMethodNotAllowed)
			return
		}

		var rec datasets.StatusGeneral8100
		err := acu.StatusGeneral8100Get(&rec)
		if err != nil {
			jsonResponse(w, 0, err, http.StatusInternalServerError)
			return
		}

		// XXX: encoding/json doesn't handle NaNs
		if math.IsNaN(rec.AzimuthCommandedPosition) {
			rec.AzimuthCommandedPosition = -1e9
		}
		if math.IsNaN(rec.ElevationCommandedPosition) {
			rec.ElevationCommandedPosition = -1e9
		}

		err = json.NewEncoder(w).Encode(&rec)
		if err != nil {
			logError(err)
		}
	})

	mux.HandleFunc("/acu/failure-reset", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			err := fmt.Errorf("method not POST")
			jsonResponse(w, 0, err, http.StatusMethodNotAllowed)
			return
		}

		id := newCmdID()
		logNewCmd(id, "acu failure reset")
		err := acu.FailureReset()
		status := http.StatusOK
		if err != nil {
			status = http.StatusInternalServerError
		}
		jsonResponse(w, id, err, status)
	})

	mux.HandleFunc("/acu/reboot", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			err := fmt.Errorf("method not POST")
			jsonResponse(w, 0, err, http.StatusMethodNotAllowed)
			return
		}

		id := newCmdID()
		logNewCmd(id, "acu reboot")
		err := acu.Reboot()
		status := http.StatusOK
		if err != nil {
			status = http.StatusInternalServerError
		}
		jsonResponse(w, id, err, status)
	})

	mux.HandleFunc("/clear-track", func(w http.ResponseWriter, req *http.Request) {
		var statusCode int
		if req.Method != "POST" {
			err := fmt.Errorf("method not POST")
			jsonResponse(w, 0, err, http.StatusMethodNotAllowed)
			return
		}
		id := newCmdID()
		logNewCmd(id, "clear program track stack")
		err := acu.ProgramTrackClear()
		if err != nil {
			logCmdError(id, err)
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		jsonResponse(w, id, err, statusCode)
	})

	mux.HandleFunc("/telescope-position", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			err := fmt.Errorf("method not GET")
			jsonResponse(w, 0, err, http.StatusMethodNotAllowed)
			return
		}
		err := json.NewEncoder(w).Encode(&tel_pos)
		if err != nil {
			logError(err)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var cmd Command
		var id uint64
		var err error
		var statusCode int

		// parse command
		if req.Method == "POST" {
			dec := json.NewDecoder(req.Body)
			dec.DisallowUnknownFields()
			endpoint := req.URL.Path
			switch endpoint {
			case "/acu/position-broadcast":
				var x enablePositionBroadcastCmd
				err = dec.Decode(&x)
				cmd = x
			case "/azimuth-scan":
				var x azScanCmd
				err = dec.Decode(&x)
				cmd = x
			case "/move-to":
				var x moveToCmd
				err = dec.Decode(&x)
				cmd = x
			case "/path":
				var x pathCmd
				err = dec.Decode(&x)
				cmd = x
			case "/track":
				var x trackCmd
				err = dec.Decode(&x)
				cmd = x
			default:
				err = fmt.Errorf("bad endpoint: %s", endpoint)
				statusCode = http.StatusNotFound
				goto respond
			}
			id = newCmdID()
			if err != nil {
				statusCode = http.StatusBadRequest
				goto respond
			}
		} else {
			// XXX:TODO: hacky
			endpoint := req.URL.Path
			switch endpoint {
			case "/azimuth-scan", "/enable-udp-stream", "/move-to", "/path", "/track":
				err = fmt.Errorf("method not POST")
				statusCode = http.StatusMethodNotAllowed
			default:
				err = fmt.Errorf("bad endpoint: %s", endpoint)
				statusCode = http.StatusNotFound
			}
			goto respond
		}

		// check parameters
		err = cmd.Check()
		if err != nil {
			statusCode = http.StatusBadRequest
			goto respond
		}

		// queue command
		select {
		case cmds <- queuedCmd{id, cmd}:
		case <-time.After(commandBusyTimeout):
			err = fmt.Errorf("busy")
			statusCode = http.StatusServiceUnavailable
			goto respond
		}

		statusCode = http.StatusOK
	respond:
		jsonResponse(w, id, err, statusCode)
	})

	// start accepting commands
	server := &http.Server{
		Addr:         apiAddr,
		Handler:      mux,
		ReadTimeout:  connectionTimeout,
		WriteTimeout: connectionTimeout,
	}
	log.Printf("listening on %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
