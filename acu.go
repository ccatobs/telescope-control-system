package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ccatp/antenna-control-unit/datasets"
)

// ACU manages communication with the ACU.
type ACU struct {
	Host   string
	client *http.Client
}

// NewACU returns a new connection to host.
func NewACU(host string) *ACU {
	return &ACU{
		Host:   host,
		client: &http.Client{},
	}
}

func (acu *ACU) do(req *http.Request) ([]byte, error) {
	resp, err := acu.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if string(b[:7]) == "Failed:" {
		return nil, fmt.Errorf(string(b))
	}
	return b, nil
}

func (acu *ACU) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	req.Host = acu.Host
	req.URL.Host = acu.Host
	req.URL.Scheme = "http"
	return req, err
}

func (acu *ACU) get(path string) ([]byte, error) {
	req, err := acu.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return acu.do(req)
}

func (acu *ACU) post(path, contentType string, body io.Reader) ([]byte, error) {
	req, err := acu.newRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return acu.do(req)
}

// GetMonitoringRecord fetches the MonitoringRecord dataset.
func (acu *ACU) GetMonitoringRecord(record *datasets.MonitoringRecord) error {
	b, err := acu.get("/Values?identifier=DataSets.MonitoringRecord&format=Binary")
	if err != nil {
		return err
	}
	r := bytes.NewBuffer(b)
	return binary.Read(r, binary.LittleEndian, record)
}

// ModeSet changes the mode.
func (acu *ACU) ModeSet(mode string) error {
	// ICD Section 9.1: "Before commanding or setting up a new mode,
	// it is best practice to set the antenna to Stop mode first."
	_, err := acu.get("/Command?identifier=Antenna.SkyAxes&command=Stop")
	if err != nil {
		return err
	}

	switch mode {
	case "Stop":
		// We already stopped, so do nothing.
	case "Preset", "ProgramTrack", "Rate", "SurvivalMode", "StarTrack", "MoonTrack", "SectorScan":
		_, err = acu.get("/Command?identifier=Antenna.SkyAxes&command=SetAzElMode&parameter=" + mode)
	default:
		err = fmt.Errorf("ModeSet: bad mode: %s", mode)
	}
	return err
}

// ProgramTrackClear clears the program track queue.
func (acu *ACU) ProgramTrackClear() error {
	_, err := acu.get("/Command?identifier=DataSets.TimePositionTransfer&command=Clear Stack")
	return err
}

// ProgramTrackGet get the current program track queue.
func (acu *ACU) ProgramTrackGet(points *[]datasets.TimePositionTransfer) error {
	b, err := acu.get("/GetPtStack")
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	slice := *points
	for {
		var point datasets.TimePositionTransfer
		err = (&point).ReadSSV(r)
		if err != nil {
			break
		}
		slice = append(slice, point)
	}
	*points = slice
	return err
}

// ProgramTrackSend appends points to the program track queue.
func (acu *ACU) ProgramTrackSend(points []datasets.TimePositionTransfer) error {
	var body bytes.Buffer
	for _, point := range points {
		point.WriteSSV(&body)
	}
	_, err := acu.post("/TrackFile?Type=File", "text/plain", &body)
	return err
}

// ShutterClose closes the shutter.
func (acu *ACU) ShutterClose() error {
	_, err := acu.get("/Command?command=SetShutter&parameter=Close")
	return err
}

// ShutterOpen opens the shutter.
func (acu *ACU) ShutterOpen() error {
	_, err := acu.get("/Command?command=SetShutter&parameter=Open")
	return err
}

// SunAvoidanceDisable disables sun avoidance.
func (acu *ACU) SunAvoidanceDisable() error {
	_, err := acu.get("/Command?command=SetSunAvoidance&parameter=Disable")
	return err
}

// SunAvoidanceEnable enables sun avoidance.
func (acu *ACU) SunAvoidanceEnable() error {
	_, err := acu.get("/Command?command=SetSunAvoidance&parameter=Enable")
	return err
}
