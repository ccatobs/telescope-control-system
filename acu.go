package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

// ACU manages communication with the ACU.
type ACU struct {
	Addr      string
	AdminAddr string
	client    *http.Client
}

// NewACU returns a new connection to host.
func NewACU(host, port, adminPort string) *ACU {
	addr := fmt.Sprintf("%s:%s", host, port)
	adminAddr := fmt.Sprintf("%s:%s", host, adminPort)
	return &ACU{
		Addr:      addr,
		AdminAddr: adminAddr,
		client: &http.Client{
			Timeout: 500 * time.Millisecond,
		},
	}
}

func (acu *ACU) do(req *http.Request) ([]byte, error) {
	resp, err := acu.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}
	if strings.HasPrefix(string(b), "Failed:") {
		return nil, fmt.Errorf(string(b))
	}
	return b, nil
}

func (acu *ACU) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	req.Host = acu.Addr
	req.URL.Host = acu.Addr
	req.URL.Scheme = "http"
	return req, err
}

func (acu *ACU) newAdminRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	req.Host = acu.AdminAddr
	req.URL.Host = acu.AdminAddr
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

func (acu *ACU) postAdminValues(path string, values url.Values) ([]byte, error) {
	body := strings.NewReader(values.Encode())
	req, err := acu.newAdminRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return acu.do(req)
}

// DatasetGet fetches a dataset.
func (acu *ACU) DatasetGet(name string, d interface{}) error {
	b, err := acu.get("/Values?identifier=DataSets." + name + "&format=Binary")
	if err != nil {
		return err
	}
	r := bytes.NewBuffer(b)
	return binary.Read(r, binary.LittleEndian, d)
}

// ModeSet changes the mode.
func (acu *ACU) ModeSet(mode string) error {
	// ICD Section 9.1: "Before commanding or setting up a new mode,
	// it is best practice to set the antenna to Stop mode first."
	_, err := acu.get("/Command?identifier=DataSets.CmdModeTransfer&command=Stop")
	if err != nil {
		return err
	}

	switch mode {
	case "Stop":
		// We already stopped, so do nothing.
	case "Preset", "ProgramTrack", "Rate", "SurvivalMode", "StarTrack", "MoonTrack", "SectorScan":
		_, err = acu.get("/Command?identifier=DataSets.CmdModeTransfer&command=SetAzElMode&parameter=" + mode)
	default:
		err = fmt.Errorf("ModeSet: bad mode: %s", mode)
	}
	return err
}

// StatusGeneral8100Get fetches the StatusGeneral8100 dataset.
func (acu *ACU) StatusGeneral8100Get(record *datasets.StatusGeneral8100) error {
	b, err := acu.get("/Values?identifier=DataSets.StatusGeneral8100&format=Binary")
	if err != nil {
		return err
	}
	r := bytes.NewBuffer(b)
	return binary.Read(r, binary.LittleEndian, record)
}

// PresetPositionSet sets the preset position.
func (acu *ACU) PresetPositionSet(azimuth, elevation float64) error {
	path := fmt.Sprintf("/Command?identifier=DataSets.CmdAzElPositionTransfer&command=Set+Azimuth+Elevation&parameter=%g|%g",
		azimuth, elevation)
	_, err := acu.get(path)
	return err
}

// ProgramTrackClear clears the program track queue.
func (acu *ACU) ProgramTrackClear() error {
	_, err := acu.get("/Command?identifier=DataSets.CmdTimePositionTransfer&command=Clear+Stack")
	return err
}

// ProgramTrackAdd appends points to the program track queue.
func (acu *ACU) ProgramTrackAdd(points []datasets.TimePositionTransfer) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("upload", "TCS")
	if err != nil {
		return err
	}
	for _, point := range points {
		point.WriteSSV(part)
	}
	part.Write([]byte("\r\n"))
	err = writer.Close()
	if err != nil {
		return err
	}
	_, err = acu.post("/UploadPtStack?type=FileMultipart", writer.FormDataContentType(), &body)
	return err
}

// ProgramTrackGet gets the current program track queue.
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

// PositionBroadcastEnable enables the 200Hz position broadcast UDP stream.
func (acu *ACU) PositionBroadcastEnable(host string, port int) error {
	data := url.Values{}
	data.Set("name", "Destination")
	data.Set("value", host)
	_, err := acu.postAdminValues("/?Module=Services.PositionBroadcast&Chapter=1", data)
	if err != nil {
		return err
	}

	data = url.Values{}
	data.Set("name", "Port")
	data.Set("value", strconv.Itoa(port))
	_, err = acu.postAdminValues("/?Module=Services.PositionBroadcast&Chapter=1", data)
	if err != nil {
		return err
	}

	data = url.Values{}
	data.Set("Command", "Enable")
	_, err = acu.postAdminValues("/?Module=Services.PositionBroadcast&Chapter=3", data)
	if err != nil {
		return err
	}

	return nil
}

// FailureReset needs to be called after an e-stop is triggered and reset.
func (acu *ACU) FailureReset() error {
	data := url.Values{}
	data.Set("Command", "Failure Reset")
	_, err := acu.postAdminValues("/?Module=DataSets.CmdGeneralTransfer&Chapter=3&Command", data)
	return err
}

// Reboot reboots the ACU.
func (acu *ACU) Reboot() error {
	data := url.Values{}
	data.Set("Command", "ACU Reboot")
	_, err := acu.postAdminValues("/?Module=DataSets.CmdGeneralTransfer&Chapter=3&Command", data)
	return err
}
