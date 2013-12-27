/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 26.12.13
 */
package translator

import (
	"selenium-hub/session"
	"encoding/json"
	"io/ioutil"
	"io"
)

type response struct {
	SessionID interface {} `json:"sessionId"`
	Status    uint8        `json:"status"`
	Value     interface {} `json:"value"`
}

type CreateSessionAnswer struct {
	SessionID string                `json:"sessionId"`
	Status    uint8                 `json:"status"`
	Value     session.Capabilities  `json:"value"`
}

type capabilities struct {
	session.Capabilities
	SeleniumProtocol string `json:"seleniumProtocol"`
	MaxInstances     uint8  `json:"maxInstances"`
}

type configuration struct {
	Port          uint16 `json:"port"`
	Host          string `json:"host"`
	MaxSession    uint8  `json:"maxSession"`
	RegisterCycle uint32 `json:"registerCycle"`
	Url           string `json:"url"`
}

type Proxy struct {
	Capabilities  []*capabilities  `json:"capabilities"`
	Configuration configuration    `json:"configuration"`
}

type apiProxyResponse struct {
	Request Proxy  `json:"request"`
	Success bool   `json:"success"`
}

type createSessionRequest struct {
	DesiredCapabilities session.Capabilities `json:"desiredCapabilities"`
}

func GetResponse(status uint8, value interface {}) ([]byte) {
	data, _ := json.Marshal(response{nil, status, value})
	return data
}

func GetProxy(reader io.Reader) (*Proxy, error) {
	machine := &Proxy{}
	setDefaults(machine)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, machine)
	if err != nil {
		return nil, err
	}
	for _, capabilities := range machine.Capabilities {
		if capabilities.MaxInstances == 0 {
			capabilities.MaxInstances = 5
		}
	}
	return machine, err
}

func setDefaults(machine *Proxy) {
	machine.Configuration.Port = 4444
	machine.Configuration.Host = "localhost"
	machine.Configuration.MaxSession = 5
	machine.Configuration.RegisterCycle = 5000
	machine.Configuration.Url = "http://localhost:4444/"
}

func GetApiProxyResponseData(machine *Proxy) ([]byte) {
	data, _ := json.Marshal(apiProxyResponse{*machine, true})
	return data
}

func GetCreateSessionCapabilities(data []byte) (*session.Capabilities, error) {
	request := createSessionRequest{}
	err := json.Unmarshal(data, &request)
	if err != nil {
		return nil, err
	}
	return &request.DesiredCapabilities, nil
}

func GetCreateSessionRequestData(capabilities *session.Capabilities) ([]byte) {
	data, _ := json.Marshal(createSessionRequest{*capabilities})
	return data
}

func GetCreateSessionAnswerData(seleniumSession *session.Session) ([]byte) {
	data, _ := json.Marshal(CreateSessionAnswer{seleniumSession.Id, 0, *seleniumSession.Capabilities})
	return data
}

func GetCreateSessionAnswer(data []byte) (*CreateSessionAnswer) {
	seleniumSession := &CreateSessionAnswer{}
	json.Unmarshal(data, seleniumSession)
	return seleniumSession
}
