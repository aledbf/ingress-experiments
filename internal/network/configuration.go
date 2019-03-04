package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/klog"

	"github.com/aledbf/ingress-experiments/internal/common"
	"github.com/aledbf/ingress-experiments/internal/nginx"
)

type ConfigurationRequest struct {
	Token      string `json:"token,omitempty"`
	LastUpdate string `json:"lastUpdate,omitempty"`
}

func RequestConfiguration(clientCfg *common.AgentConfiguration) (*nginx.Configuration, bool) {
	request := ConfigurationRequest{
		Token:      "",
		LastUpdate: "",
	}

	var response nginx.Configuration

	result, _, err := newUpdateRequest(request)
	if err != nil {
		return nil, false
	}

	klog.Infof("Checking for jobs (%v)", http.StatusCreated)

	switch result {
	case http.StatusCreated:
		return &response, true
	case http.StatusForbidden:
		return nil, false
	case http.StatusNoContent:
		return nil, true
	default:
		return nil, true
	}
}

func newUpdateRequest(data interface{}) (int, []byte, error) {
	client := &http.Client{}

	url := fmt.Sprintf("http://%v/v1/check-update", "")

	buf, err := json.Marshal(data)
	if err != nil {
		return 0, nil, err
	}

	res, err := client.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	return res.StatusCode, body, nil
}
