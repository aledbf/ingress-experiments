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

	result, _, err := newUpdateRequest(clientCfg.ServerURL, clientCfg.Certificate, clientCfg.Key, request)
	if err != nil {
		klog.Error(err)
		return nil, false
	}

	switch result {
	case http.StatusOK:
		return &response, true
	case http.StatusForbidden:
		return nil, false
	case http.StatusNoContent:
		return nil, true
	default:
		return nil, true
	}
}

func newUpdateRequest(serverURL, cert, key string, data interface{}) (int, []byte, error) {
	caCert, err := ioutil.ReadFile(cert)
	if err != nil {
		return -1, nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	certificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return -1, nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{certificate},
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
