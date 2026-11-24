package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"math"
	"net/http"
	"os"
	"slices"
	"strings"
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

func logNewCmd(id uint64, desc string, tags Tags) {
	if len(desc) > 200 {
		desc = fmt.Sprintf("%.200s...", desc)
	}
	slog.Info("got command", "cmd", id, "desc", desc, "tags", tags)
}

func logCmdError(id uint64, err error) {
	if err != nil {
		slog.Error(err.Error(), "cmd", id)
	}
}

// Command IDs come from a simple incrementing counter. It's seeded with
// the startup time in milliseconds since cmdIDEpoch, so IDs stay unique
// across restarts. A custom epoch keeps them from looking like Unix timestamps.
var lastCmdID atomic.Uint64

var cmdIDEpoch = time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC) // FYST inauguration

func init() {
	lastCmdID.Store(uint64(time.Since(cmdIDEpoch).Milliseconds()))
}

func newCmdID() uint64 {
	return lastCmdID.Add(1)
}

// Tags are user-supplied key/value pairs attached to a command.
type Tags map[string]string

// limits on user-supplied tags
const (
	maxTags        = 16
	maxTagKeyLen   = 64
	maxTagValueLen = 256
)

func (tags Tags) Check() error {
	if len(tags) > maxTags {
		return fmt.Errorf("too many tags (%d > %d)", len(tags), maxTags)
	}
	for k, v := range tags {
		if k == "" {
			return fmt.Errorf("empty tag key")
		}
		if len(k) > maxTagKeyLen {
			return fmt.Errorf("tag key %.20q... too long (%d > %d bytes)", k, len(k), maxTagKeyLen)
		}
		if len(v) > maxTagValueLen {
			return fmt.Errorf("tag %q value too long (%d > %d bytes)", k, len(v), maxTagValueLen)
		}
	}
	return nil
}

// LogValue logs tags as a group, e.g. tags.key=val, sorted by key.
func (tags Tags) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(tags))
	for k, v := range tags {
		attrs = append(attrs, slog.String(k, v))
	}
	slices.SortFunc(attrs, func(a, b slog.Attr) int { return strings.Compare(a.Key, b.Key) })
	return slog.GroupValue(attrs...)
}

// decodeCmdBody decodes a JSON command body into v, rejecting unknown fields,
// and returns the tags from its optional "tags" field.
func decodeCmdBody(body io.Reader, v any) (Tags, error) {
	var fields map[string]json.RawMessage
	err := json.NewDecoder(body).Decode(&fields)
	if err != nil {
		return nil, err
	}

	var tags Tags
	if raw, ok := fields["tags"]; ok {
		err = json.Unmarshal(raw, &tags)
		if err == nil {
			err = tags.Check()
		}
		if err != nil {
			return nil, fmt.Errorf("bad tags: %w", err)
		}
		delete(fields, "tags")
	}

	rest, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(rest))
	dec.DisallowUnknownFields()
	return tags, dec.Decode(v)
}

// checkNoBody returns an error if a command which takes no parameters has a body.
func checkNoBody(body io.Reader) error {
	b, err := io.ReadAll(io.LimitReader(body, 1))
	if err != nil {
		return err
	}
	if len(b) > 0 {
		return fmt.Errorf("command takes no parameters")
	}
	return nil
}

// cmdParser parses a request body into a Command & its user tags.
type cmdParser func(body io.Reader) (Command, Tags, error)

// paramsCmd is a cmdParser for commands taking JSON parameters.
func paramsCmd[T Command](body io.Reader) (Command, Tags, error) {
	var x T
	tags, err := decodeCmdBody(body, &x)
	return x, tags, err
}

// noParamsCmd returns a cmdParser for commands taking no parameters.
func noParamsCmd(cmd Command) cmdParser {
	return func(body io.Reader) (Command, Tags, error) {
		return cmd, nil, checkNoBody(body)
	}
}

// queuedCmd is a Command tagged with its ID, endpoint, and user tags.
type queuedCmd struct {
	id       uint64
	endpoint string
	tags     Tags
	cmd      Command
}

// runningCmd describes the currently running command, as reported by /status.
type runningCmd struct {
	ID        uint64  `json:"id"`
	Endpoint  string  `json:"endpoint"`
	Params    any     `json:"params"`
	Tags      Tags    `json:"tags,omitempty"`
	StartedAt float64 `json:"started_at"`
}

// statusParams returns the params of cmd to report in /status.
// Paths can be long, so their points are replaced by a count.
func statusParams(cmd Command) any {
	if x, ok := cmd.(pathCmd); ok {
		return struct {
			Coordsys  string  `json:"coordsys"`
			NumPoints int     `json:"num_points"`
			StartTime float64 `json:"start_time"`
		}{x.Coordsys, len(x.Points), x.StartTime}
	}
	return cmd
}

// statusResponse writes the /status response. A nil rc means no command is running.
func statusResponse(w http.ResponseWriter, rc *runningCmd) {
	response := struct {
		S       string      `json:"status"`
		Command *runningCmd `json:"command"`
	}{"ok", rc}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(response)
	if err != nil {
		logError(err)
	}
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
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err = enc.Encode(response)
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
	trackUploadURL := os.Getenv("TCS_PROGRAM_TRACK_UPLOAD_URL")

	acu := NewACU(acuHost, acuPort, acuMonitorPort, acuAdminPort)
	tel := NewTelescope(acu)
	tel.trackUploadURL = trackUploadURL

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

	// currently running command, or nil
	var running atomic.Pointer[runningCmd]

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
					slog.Info("ignoring abort: no command running")
					c <- false
				}
			}

			cmd := qc.cmd
			logNewCmd(qc.id, fmt.Sprintf("%#v", cmd), qc.tags)

			if err := tel.Ready(); err != nil {
				logCmdError(qc.id, err)
				continue
			}

			// start command
			running.Store(&runningCmd{
				ID:        qc.id,
				Endpoint:  qc.endpoint,
				Params:    statusParams(cmd),
				Tags:      qc.tags,
				StartedAt: Time2Unixtime(time.Now()),
			})
			ctx, cancel := context.WithCancel(context.Background())
			isDone, err := cmd.Start(ctx, tel)
			if err != nil {
				logCmdError(qc.id, err)
				cancel()
				tel.WaitForUploads()
				tel.EndProgramTrack(false)
				running.Store(nil)
				continue
			}

			// wait for command to finish
			aborted := false
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
					aborted = true
					cancel()
					err = tel.Stop()
				}
				if err != nil {
					logCmdError(qc.id, err)
					break
				}
			}
			// stop any background work (e.g. pattern uploads), however the command ended
			cancel()
			tel.WaitForUploads()
			// after WaitForUploads, so no batch is posted after this
			tel.EndProgramTrack(!aborted && err == nil)

			running.Store(nil)
			slog.Info("command done", "cmd", qc.id)
		}
	}()

	// build http API
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, req *http.Request) {
		err := tel.Ready()
		statusCode := http.StatusServiceUnavailable
		if err == nil {
			statusCode = http.StatusOK
		}
		jsonResponse(w, 0, err, statusCode)
	})

	mux.HandleFunc("GET /status", func(w http.ResponseWriter, req *http.Request) {
		statusResponse(w, running.Load())
	})

	mux.HandleFunc("POST /abort", func(w http.ResponseWriter, req *http.Request) {
		id := newCmdID()
		err := checkNoBody(req.Body)
		if err != nil {
			jsonResponse(w, id, err, http.StatusBadRequest)
			return
		}
		logNewCmd(id, "abort", nil)
		c := make(chan bool)
		abort <- c
		statusCode := http.StatusOK
		if !<-c {
			err = fmt.Errorf("nothing to abort")
			statusCode = http.StatusConflict // not sure if this is the most appropriate code
		}
		jsonResponse(w, id, err, statusCode)
	})

	mux.HandleFunc("POST /reset", func(w http.ResponseWriter, req *http.Request) {
		id := newCmdID()
		err := checkNoBody(req.Body)
		if err != nil {
			jsonResponse(w, id, err, http.StatusBadRequest)
			return
		}
		logNewCmd(id, "reset", nil)
		c := make(chan bool)
		abort <- c
		if <-c {
			slog.Info("aborted running command", "cmd", id)
		}
		// stop before resetting failures, so the axes don't resume a previous mode;
		// try both, even if stopping fails (e.g. because of a latched failure)
		err = errors.Join(tel.Stop(), acu.FailureReset())
		statusCode := http.StatusOK
		if err != nil {
			logCmdError(id, err)
			statusCode = http.StatusInternalServerError
		}
		jsonResponse(w, id, err, statusCode)
	})

	mux.HandleFunc("GET /acu/status", func(w http.ResponseWriter, req *http.Request) {
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

	mux.HandleFunc("GET /acu/datasets", func(w http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(w).Encode(slices.Sorted(maps.Keys(acuDatasets)))
		if err != nil {
			logError(err)
		}
	})

	mux.HandleFunc("GET /acu/datasets/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := req.PathValue("name")
		newRec, ok := acuDatasets[name]
		if !ok {
			err := fmt.Errorf("unknown dataset: %s", name)
			jsonResponse(w, 0, err, http.StatusNotFound)
			return
		}

		rec := newRec()
		err := acu.DatasetGet(name, rec)
		if err != nil {
			jsonResponse(w, 0, err, http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(rec)
		if err != nil {
			jsonResponse(w, 0, err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(append(b, '\n'))
	})

	mux.HandleFunc("GET /acu/values/{identifier}", func(w http.ResponseWriter, req *http.Request) {
		// forward the identifier as the client escaped it, so that
		// the ACU decodes it (e.g. "+" and "%20" are both spaces)
		id := strings.TrimPrefix(req.URL.EscapedPath(), "/acu/values/")
		b, err := acu.ValuesGet(id)
		if err != nil {
			jsonResponse(w, 0, err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	mux.HandleFunc("POST /acu/failure-reset", func(w http.ResponseWriter, req *http.Request) {
		id := newCmdID()
		err := checkNoBody(req.Body)
		if err != nil {
			jsonResponse(w, id, err, http.StatusBadRequest)
			return
		}
		logNewCmd(id, "acu failure reset", nil)
		err = acu.FailureReset()
		status := http.StatusOK
		if err != nil {
			status = http.StatusInternalServerError
		}
		jsonResponse(w, id, err, status)
	})

	mux.HandleFunc("POST /acu/reboot", func(w http.ResponseWriter, req *http.Request) {
		id := newCmdID()
		err := checkNoBody(req.Body)
		if err != nil {
			jsonResponse(w, id, err, http.StatusBadRequest)
			return
		}
		logNewCmd(id, "acu reboot", nil)
		err = acu.Reboot()
		status := http.StatusOK
		if err != nil {
			status = http.StatusInternalServerError
		}
		jsonResponse(w, id, err, status)
	})

	// deprecated wraps a handler for an old endpoint name, logging a warning.
	deprecated := func(newEndpoint string, h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			slog.Warn("deprecated endpoint", "endpoint", req.URL.Path, "use", newEndpoint)
			h(w, req)
		}
	}

	clearPath := func(w http.ResponseWriter, req *http.Request) {
		id := newCmdID()
		err := checkNoBody(req.Body)
		if err != nil {
			jsonResponse(w, id, err, http.StatusBadRequest)
			return
		}
		logNewCmd(id, "clear program track stack", nil)
		err = tel.ClearProgramTrack()
		statusCode := http.StatusOK
		if err != nil {
			logCmdError(id, err)
			statusCode = http.StatusBadRequest
		}
		jsonResponse(w, id, err, statusCode)
	}
	mux.HandleFunc("POST /path/clear", clearPath)
	mux.HandleFunc("POST /clear-track", deprecated("/path/clear", clearPath)) // deprecated

	mux.HandleFunc("GET /telescope-position", func(w http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(w).Encode(&tel_pos)
		if err != nil {
			logError(err)
		}
	})

	// queued commands
	commandHandler := func(parse cmdParser) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			cmd, tags, err := parse(req.Body)
			id := newCmdID()
			if err == nil {
				err = cmd.Check()
			}
			if err != nil {
				jsonResponse(w, id, err, http.StatusBadRequest)
				return
			}

			// stow & survival mode take priority over the running command
			switch cmd.(type) {
			case stowCmd, survivalModeCmd:
				c := make(chan bool)
				abort <- c
				if <-c {
					slog.Info("aborted running command", "cmd", id)
				}
			}

			// queue command
			select {
			case cmds <- queuedCmd{id, req.URL.Path, tags, cmd}:
			case <-time.After(commandBusyTimeout):
				jsonResponse(w, id, fmt.Errorf("busy"), http.StatusServiceUnavailable)
				return
			}
			jsonResponse(w, id, nil, http.StatusOK)
		}
	}
	mux.HandleFunc("POST /acu/position-broadcast", commandHandler(paramsCmd[enablePositionBroadcastCmd]))
	mux.HandleFunc("POST /acu/spem", commandHandler(paramsCmd[setSPEMCmd]))
	mux.HandleFunc("POST /acu/stow", commandHandler(noParamsCmd(stowCmd{})))
	mux.HandleFunc("POST /acu/survival-mode", commandHandler(noParamsCmd(survivalModeCmd{})))
	mux.HandleFunc("POST /acu/unstow", commandHandler(noParamsCmd(unstowCmd{})))
	mux.HandleFunc("POST /azimuth-scan", commandHandler(paramsCmd[azScanCmd]))
	mux.HandleFunc("POST /holography-scan", commandHandler(paramsCmd[holographyScanCmd]))
	mux.HandleFunc("POST /move-to", commandHandler(paramsCmd[moveToCmd]))
	mux.HandleFunc("POST /path", commandHandler(paramsCmd[pathCmd]))
	mux.HandleFunc("POST /stow", deprecated("/acu/stow", commandHandler(noParamsCmd(stowCmd{}))))                           // deprecated
	mux.HandleFunc("POST /survival-mode", deprecated("/acu/survival-mode", commandHandler(noParamsCmd(survivalModeCmd{})))) // deprecated
	mux.HandleFunc("POST /track", commandHandler(paramsCmd[trackCmd]))
	mux.HandleFunc("POST /unstow", deprecated("/acu/unstow", commandHandler(noParamsCmd(unstowCmd{})))) // deprecated

	// catch-all: JSON 404 & 405 responses
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var allowed []string
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			r := req.Clone(req.Context())
			r.Method = method
			if _, pattern := mux.Handler(r); pattern != "/" {
				allowed = append(allowed, method)
			}
		}
		if len(allowed) > 0 {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			err := fmt.Errorf("method not %s", strings.Join(allowed, " or "))
			jsonResponse(w, 0, err, http.StatusMethodNotAllowed)
			return
		}
		jsonResponse(w, 0, fmt.Errorf("bad endpoint: %s", req.URL.Path), http.StatusNotFound)
	})

	// start accepting commands
	server := &http.Server{
		Addr:         apiAddr,
		Handler:      mux,
		ReadTimeout:  connectionTimeout,
		WriteTimeout: connectionTimeout,
	}
	slog.Info("listening", "addr", server.Addr)
	err = server.ListenAndServe()
	slog.Error(err.Error())
	os.Exit(1)
}
