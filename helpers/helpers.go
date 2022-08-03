package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type ConfigJsonData struct {
	AppiumConfig AppiumConfig   `json:"appium-config"`
	EnvConfig    EnvConfig      `json:"env-config"`
	DeviceConfig []DeviceConfig `json:"devices-config"`
}

type AppiumConfig struct {
	DevicesHost             string `json:"devices_host"`
	SeleniumHubHost         string `json:"selenium_hub_host"`
	SeleniumHubPort         string `json:"selenium_hub_port"`
	SeleniumHubProtocolType string `json:"selenium_hub_protocol_type"`
	WDABundleID             string `json:"wda_bundle_id"`
}

type EnvConfig struct {
	ConnectSeleniumGrid  string `json:"connect_selenium_grid"`
	SupervisionPassword  string `json:"supervision_password"`
	ContainerizedUsbmuxd string `json:"containerized_usbmuxd"`
	RemoteControl        string `json:"remote_control"`
}

type DeviceConfig struct {
	OS                  string `json:"os"`
	AppiumPort          string `json:"appium_port"`
	DeviceName          string `json:"device_name"`
	DeviceOSVersion     string `json:"device_os_version"`
	DeviceUDID          string `json:"device_udid"`
	StreamPort          string `json:"stream_port"`
	WDAPort             string `json:"wda_port"`
	ScreenSize          string `json:"screen_size"`
	ContainerServerPort string `json:"container_server_port"`
	DeviceModel         string `json:"device_model"`
	DeviceImage         string `json:"device_image"`
	DeviceHost          string `json:"device_host"`
}

type JsonErrorResponse struct {
	EventName    string `json:"event"`
	ErrorMessage string `json:"error_message"`
}

type JsonResponse struct {
	Message string `json:"message"`
}

var ProviderPort, ProviderPath string

func ValidateFlags(provider_port string, provider_path string) error {
	ProviderPort = provider_port
	ProviderPath = provider_path

	if ProviderPath == "" {
		return errors.New("Please provide an absolute path for the provider logs/files without a trailing slash. Example: -provider_path=/home/shamanec/gads-provider")
	}

	return nil
}

//=======================//
//=====API FUNCTIONS=====//

// Write to a ResponseWriter an event and message with a response code
func JSONError(w http.ResponseWriter, event string, error_string string, code int) {
	var errorMessage = JsonErrorResponse{
		EventName:    event,
		ErrorMessage: error_string}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorMessage)
}

// Write to a ResponseWriter an event and message with a response code
func SimpleJSONResponse(w http.ResponseWriter, response_message string, code int) {
	var message = JsonResponse{
		Message: response_message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(message)
}

// @Summary      Get provider logs
// @Description  Gets provider logs as plain text response
// @Tags         provider-logs
// @Produces	 text
// @Success      200
// @Failure      200
// @Router       /provider-logs [get]
func GetLogs(w http.ResponseWriter, r *http.Request) {
	// Create the command string to read the last 1000 lines of provider.log
	commandString := "tail -n 1000 " + ProviderPath + "/provider.log"

	// Create the command
	cmd := exec.Command("bash", "-c", commandString)

	// Create a buffer for the output
	var out bytes.Buffer

	// Pipe the Stdout of the command to the buffer pointer
	cmd.Stdout = &out

	// Execute the command
	err := cmd.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"event": "get_project_logs",
		}).Warning("Attempted to get project logs but no logs available.")

		// Reply with generic message on error
		fmt.Fprintf(w, "No logs available.")
		return
	}

	// Reply with the read logs lines
	fmt.Fprintf(w, out.String())
}

//=======================//
//=======FUNCTIONS=======//

// Get a ConfigJsonData pointer with the current configuration from config.json
func GetConfigJsonData() (*ConfigJsonData, error) {
	var data ConfigJsonData
	jsonFile, err := os.Open(ProviderPath + "/config.json")
	if err != nil {
		log.WithFields(log.Fields{
			"event": "get_config_data",
		}).Error("Could not open config file: " + err.Error())
		return nil, err
	}
	defer jsonFile.Close()

	bs, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "get_config_data",
		}).Error("Could not read config file to byte slice: " + err.Error())
		return nil, err
	}

	err = json.Unmarshal(bs, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "get_config_data",
		}).Error("Could not unmarshal config file: " + err.Error())
		return nil, err
	}

	return &data, nil
}

// Convert interface into JSON string
func ConvertToJSONString(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.WithFields(log.Fields{
			"event": "convert_interface_to_json",
		}).Error("Could not marshal interface to json: " + err.Error())
		return "", err
	}
	return string(b), nil
}

// Unmarshal request body into a struct
func UnmarshalReader(body io.ReadCloser, v interface{}) error {
	reqBody, err := ioutil.ReadAll(body)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "unmarshal_reader",
		}).Error("Could not read reader into byte slice: " + err.Error())
		return err
	}

	err = json.Unmarshal(reqBody, v)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "unmarshal_reader",
		}).Error("Could not unmarshal reader: " + err.Error())
		return err
	}

	return nil
}

// Get an env value from config.json
func GetEnvValue(key string) string {
	configData, err := GetConfigJsonData()
	if err != nil {
		return ""
	}

	if key == "supervision_password" {
		return configData.EnvConfig.SupervisionPassword
	} else if key == "connect_selenium_grid" {
		return configData.EnvConfig.ConnectSeleniumGrid
	} else {
		return ""
	}
}
