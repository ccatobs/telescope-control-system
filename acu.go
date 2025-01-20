package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

const (
	minDurationBetweenCommands = 10 * time.Millisecond

	clientTimeout = 2000 * time.Millisecond
)

// ACU manages communication with the ACU.
type ACU struct {
	Addr        string
	AdminAddr   string
	client      *http.Client
	lastCommand time.Time
}

// NewACU returns a new connection to host.
func NewACU(host, port, adminPort string) *ACU {
	addr := fmt.Sprintf("%s:%s", host, port)
	adminAddr := fmt.Sprintf("%s:%s", host, adminPort)
	return &ACU{
		Addr:      addr,
		AdminAddr: adminAddr,
		client: &http.Client{
			Timeout: clientTimeout,
		},
	}
}

func (acu *ACU) do(req *http.Request) ([]byte, error) {
	resp, err := acu.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
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
	if !strings.HasPrefix(path, "/Values") { // cut down on log spam
		log.Printf("ACU: GET %s", path)
	}
	req, err := acu.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return acu.do(req)
}

func (acu *ACU) post(path, contentType string, body io.Reader) ([]byte, error) {
	log.Printf("ACU: POST %s", path)
	req, err := acu.newRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	acu.commandWait()
	return acu.do(req)
}

func (acu *ACU) postAdminValues(path string, values url.Values) ([]byte, error) {
	body := strings.NewReader(values.Encode())
	req, err := acu.newAdminRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	acu.commandWait()
	return acu.do(req)
}

func (acu *ACU) commandWait() {
	// don't send commands too quickly
	dt := minDurationBetweenCommands - time.Since(acu.lastCommand)
	if dt > 0 {
		slog.Debug("sendCommand", "sleep", dt)
		time.Sleep(dt)
	}
	acu.lastCommand = time.Now()
}

func (acu *ACU) sendCommand(url string) error {
	acu.commandWait()
	_, err := acu.get(url)
	return err
}

func (acu *ACU) command2(id, cmd string) error {
	return acu.sendCommand("/Command?identifier=" + id + "&command=" + cmd)
}

func (acu *ACU) command3(id, cmd, param string) error {
	return acu.sendCommand("/Command?identifier=" + id + "&command=" + cmd + "&parameter=" + param)
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
	switch mode {
	case "Stop":
		return acu.command2("DataSets.CmdModeTransfer", "Stop")
	case "Preset", "ProgramTrack", "Rate", "SurvivalMode", "StarTrack", "MoonTrack", "SectorScan":
		return acu.command3("DataSets.CmdModeTransfer", "SetAzElMode", mode)
	}
	return fmt.Errorf("ModeSet: bad mode: %s", mode)
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
	return acu.command3("DataSets.CmdAzElPositionTransfer", "Set+Azimuth+Elevation", fmt.Sprintf("%g|%g", azimuth, elevation))
}

// ProgramTrackClear clears the program track queue.
func (acu *ACU) ProgramTrackClear() error {
	return acu.command2("DataSets.CmdTimePositionTransfer", "Clear+Stack")
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
	if err != nil {
		return err
	}
	var details datasets.StatusCCatDetailed8100
	err = acu.DatasetGet("StatusCCatDetailed8100", &details)
	if err != nil {
		return err
	}
	if details.StartOfProgramTrackTooEarly {
		return fmt.Errorf("ProgramTrackAdd: StartOfProgramTrackTooEarly")
	}
	if details.ProgramTrackPositionFailure {
		return fmt.Errorf("ProgramTrackAdd: ProgramTrackPositionFailure")
	}
	return nil
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
	return acu.command2("SetShutter", "Close")
}

// ShutterOpen opens the shutter.
func (acu *ACU) ShutterOpen() error {
	return acu.command2("SetShutter", "Open")
}

// SunAvoidanceDisable disables sun avoidance.
func (acu *ACU) SunAvoidanceDisable() error {
	return acu.command2("SetSunAvoidance", "Disable")
}

// SunAvoidanceEnable enables sun avoidance.
func (acu *ACU) SunAvoidanceEnable() error {
	return acu.command2("SetSunAvoidance", "Enable")
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
	return acu.command2("DataSets.CmdGeneralTransfer", "Failure+Reset")
}

// Reboot reboots the ACU.
func (acu *ACU) Reboot() error {
	return acu.command2("DataSets.CmdGeneralTransfer", "ACU+Reboot")
}
