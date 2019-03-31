package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ccatp/antenna-control-unit/datasets"
)

const (
	// poll the ACU status at 1 Hz
	acuPollDuration = 1000 * time.Millisecond

	// max time waiting to queue command
	commandBusyTimeout = 100 * time.Millisecond

	// http connection timeout
	connectionTimeout = 1000 * time.Millisecond
)

func init() {
	// log format: date time(UTC) file:linenumber message
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)
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

func main() {
	var acuAddr string
	var apiAddr string
	flag.StringVar(&acuAddr, "acu", "172.16.5.95:8100", "ACU address")
	flag.StringVar(&apiAddr, "api", ":5600", "HTTP API address")
	flag.Parse()

	acu := NewACU(acuAddr)
	tel := NewTelescope(acu)

	// XXX:DEBUG fake pointing model
	tel.pointing.azOffset = 5
	tel.pointing.elOffset = 6

	// command queue
	cmds := make(chan Command)

	// abort signal
	abort := make(chan chan bool)

	// main loop
	go func() {
		var rec datasets.MonitoringRecord
		for {
			// wait for command
			var cmd Command
		waitForCmdLoop:
			for {
				select {
				case cmd = <-cmds:
					break waitForCmdLoop
				case <-time.After(acuPollDuration):
					err := acu.MonitoringRecordGet(&rec)
					if err != nil {
						log.Print(err)
					}
				case c := <-abort:
					log.Print("ignoring abort")
					c <- false
				}
			}

			desc := fmt.Sprintf("%#v", cmd)
			log.Printf("got command: %s", desc)

			if !rec.Remote {
				log.Print("ignoring command, ACU not in remote mode")
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
				case <-time.After(acuPollDuration):
					err = acu.MonitoringRecordGet(&rec)
					if err == nil {
						done, err = isDone(&rec)
					}
				case c := <-abort:
					log.Print("aborting")
					c <- true
					done = true
					cancel()
					err = tel.Stop()
				}
				if err != nil {
					log.Print(err)
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
			case "/azimuth-scan":
				var x azScanCmd
				err = dec.Decode(&x)
				cmd = x
			case "/move-to":
				var x moveToCmd
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
			err = fmt.Errorf("method not POST")
			statusCode = http.StatusMethodNotAllowed
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
