package network

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
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

	result, _, err := newUpdateRequest(clientCfg.ServerURL, clientCfg.Certificate, request)
	if err != nil {
		klog.Error(err)
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

func newUpdateRequest(serverURL, cert string, data interface{}) (int, []byte, error) {
	caCert, err := ioutil.ReadFile("cert.pem")
	if err != nil {
		return -1, nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	url := fmt.Sprintf("%v/v1/check-update", serverURL)
	klog.V(2).Infof("Server update URL: %v", url)

	buf, err := json.Marshal(data)
	if err != nil {
		return 0, nil, err
	}

	res, err := client.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return 0, nil, fmt.Errorf("ingress-nginx server returned 404. Please check %v is correct", serverURL)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	klog.Infof("%v", res)
	return res.StatusCode, body, nil
}
