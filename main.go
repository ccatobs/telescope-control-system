package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

const (
	// poll the ACU status at 1 Hz
	statusUpdateDuration = 1000 * time.Millisecond

	// max time waiting to queue command
	commandBusyTimeout = 100 * time.Millisecond

	// http connection timeout
	connectionTimeout = 1000 * time.Millisecond
)

func init() {
	// log format: date time(UTC) file:linenumber message
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC | log.Lshortfile)
}

func jsonResponse(w http.ResponseWriter, err error, statusCode int) {
	var response struct {
		S string `json:"status"`
		M string `json:"message,omitempty"`
	}

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
		log.Print(err)
	}
}

func getenv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func main() {
	acuHost := getenv("FYST_ACU_HOST", "172.16.5.95")
	acuPort := getenv("FYST_ACU_PORT", "8100")
	acuAdminPort := getenv("FYST_ACU_ADMIN_PORT", "8080")
	apiAddr := getenv("FYST_TCS_ADDR", ":5600")

	acu := NewACU(acuHost, acuPort, acuAdminPort)
	tel := NewTelescope(acu)

	// report immediately any ACU problems
	err := tel.UpdateStatus()
	if err != nil {
		log.Print(err)
	}
	err = tel.Ready()
	if err != nil {
		log.Print(err)
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
	cmds := make(chan Command)

	// abort signal
	abort := make(chan chan bool)

	// main loop
	go func() {
		for {
			// wait for command
			var cmd Command
		waitForCmdLoop:
			for {
				select {
				case cmd = <-cmds:
					break waitForCmdLoop
				case <-time.After(statusUpdateDuration):
					err := tel.UpdateStatus()
					if err != nil {
						log.Print(err)
					}
				case c := <-abort:
					log.Print("ignoring abort")
					c <- false
				}
			}

			desc := fmt.Sprintf("%#v", cmd)
			if len(desc) > 200 {
				desc = fmt.Sprintf("%.200s...", desc)
			}
			log.Printf("got command: %s", desc)

			if err := tel.Ready(); err != nil {
				log.Print(err)
				continue
			}

			// start command
			ctx, cancel := context.WithCancel(context.Background())
			isDone, err := cmd.Start(ctx, tel)
			if err != nil {
				log.Print(err)
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
					log.Print("aborting")
					c <- true
					done = true
					cancel()
					err = tel.Stop()
				}
				if err != nil {
					log.Print(err)
					break
				}
			}

			log.Printf("command done: %s", desc)
		}
	}()

	// build http API
	mux := http.NewServeMux()

	mux.HandleFunc("/abort", func(w http.ResponseWriter, req *http.Request) {
		var err error
		var statusCode int

		if req.Method == "POST" {
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

		jsonResponse(w, err, statusCode)
	})

	mux.HandleFunc("/acu/status", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			err := fmt.Errorf("method not GET")
			jsonResponse(w, err, http.StatusMethodNotAllowed)
			return
		}

		var rec datasets.StatusGeneral8100
		err := acu.StatusGeneral8100Get(&rec)
		if err != nil {
			jsonResponse(w, err, http.StatusInternalServerError)
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
			log.Print(err)
		}
	})

	mux.HandleFunc("/acu/failure-reset", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			err := fmt.Errorf("method not POST")
			jsonResponse(w, err, http.StatusMethodNotAllowed)
			return
		}

		err := acu.FailureReset()
		if err != nil {
			jsonResponse(w, err, http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/acu/reboot", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			err := fmt.Errorf("method not POST")
			jsonResponse(w, err, http.StatusMethodNotAllowed)
			return
		}

		err := acu.Reboot()
		if err != nil {
			jsonResponse(w, err, http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/clear-track", func(w http.ResponseWriter, req *http.Request) {
		var statusCode int
		if req.Method != "POST" {
			err := fmt.Errorf("method not POST")
			jsonResponse(w, err, http.StatusMethodNotAllowed)
			return
		}
		log.Print("clearing program track stack")
		err := acu.ProgramTrackClear()
		if err != nil {
			log.Print(err)
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		jsonResponse(w, err, statusCode)
	})

	mux.HandleFunc("/telescope-position", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			err := fmt.Errorf("method not GET")
			jsonResponse(w, err, http.StatusMethodNotAllowed)
			return
		}
		err := json.NewEncoder(w).Encode(&tel_pos)
		if err != nil {
			log.Print(err)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var cmd Command
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
		case cmds <- cmd:
		case <-time.After(commandBusyTimeout):
			err = fmt.Errorf("busy")
			statusCode = http.StatusServiceUnavailable
			goto respond
		}

		statusCode = http.StatusOK
	respond:
		jsonResponse(w, err, statusCode)
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
