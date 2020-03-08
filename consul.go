// Copyright 2020 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"

	"github.com/Jeffail/gabs/v2"
	"github.com/mitchellh/mapstructure"
)

type config struct {
	Provider              string
	Path                  string
	Url                   string
	Payload               string
	Token                 string
	PassingOnly           bool
	ReplaceExistingChecks bool
	InsecureSkipVerify    bool
}

type serviceDiscovery struct {
	service string
	id      string
	config  *config
	log     *log.Logger
	client  *http.Client
}

func (s *serviceDiscovery) checkErrors() error {
	if s.config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if s.log == nil {
		return fmt.Errorf("logger cannot be nil")
	}
	return nil
}

// SetConfig registers plugin configuration
func (s *serviceDiscovery) SetConfig(c map[string]interface{}) error {
	var cfg config
	err := mapstructure.Decode(c, &cfg)
	if err != nil {
		return err
	}
	s.config = &cfg
	return nil
}

// SetLogger sets an appropriate
func (s *serviceDiscovery) SetLogger(l *log.Logger) {
	s.log = l
}

// Initialize initializes the plugin: registers some internal data structures, clients etc.
func (s *serviceDiscovery) Initialize() error {
	if err := s.checkErrors(); err != nil {
		return err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: s.config.InsecureSkipVerify},
	}
	s.client = &http.Client{Transport: tr}
	s.log.Printf("[INFO] Service discovery plugin is enabled, provider: %s", s.config.Provider)
	return nil
}

func (s *serviceDiscovery) setServiceAndId() error {
	payload := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s.config.Payload), &payload); err != nil {
		return err
	}
	name, ok := payload["Name"]
	if !ok {
		return fmt.Errorf("required field not found: Name")
	}
	s.service = name.(string)

	id, ok := payload["ID"]
	if !ok {
		return fmt.Errorf("required field not found: ID")
	}
	s.service = name.(string)
	s.id = id.(string)
	return nil
}

func (s *serviceDiscovery) doRequest(op string, req *http.Request) error {
	if s.config.Token != "" {
		req.Header.Set("X-Consul-Token", s.config.Token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", op, string(data))
	}
	return nil
}

// Register registers this node to a service discovery directory.
func (s *serviceDiscovery) Register() error {
	if err := s.checkErrors(); err != nil {
		return err
	}
	u, err := url.Parse(s.config.Url)
	if err != nil {
		return nil
	}
	u.Path = path.Join(u.Path, "/v1/agent/service/register")
	if s.config.ReplaceExistingChecks {
		q := u.Query()
		q.Set("replace-existing-checks", "true")
		u.RawQuery = q.Encode()
	}

	if err := s.setServiceAndId(); err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte(s.config.Payload))
	req, err := http.NewRequest(http.MethodPut, u.String(), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return s.doRequest("register", req)
}

// Deregister removes this node from a service discovery directory.
func (s *serviceDiscovery) Deregister() error {
	u, err := url.Parse(s.config.Url)
	if err != nil {
		return nil
	}
	u.Path = path.Join(u.Path, "/v1/agent/service/deregister", s.id)
	req, err := http.NewRequest(http.MethodPut, u.String(), nil)
	if err != nil {
		return err
	}
	return s.doRequest("deregister", req)
}

// DiscoverPeers returns a list of known Olric nodes.
func (s *serviceDiscovery) DiscoverPeers() ([]string, error) {
	if err := s.checkErrors(); err != nil {
		return nil, err
	}

	u, err := url.Parse(s.config.Url)
	if err != nil {
		return nil, nil
	}
	u.Path = path.Join(u.Path, "/v1/health/service", s.service)
	if s.config.PassingOnly {
		q := u.Query()
		q.Set("passing", "1")
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parsed, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		return nil, err
	}
	var peers []string
	for _, node := range parsed.Children() {
		id, ok := node.Path("Service.ID").Data().(string)
		if !ok {
			continue
		}
		if id == s.id {
			continue
		}

		addr, ok := node.Path("Service.Address").Data().(string)
		if !ok {
			continue
		}
		port, ok := node.Path("Service.Port").Data().(float64)
		if !ok {
			continue
		}
		peer := net.JoinHostPort(addr, fmt.Sprintf("%v", port))
		peers = append(peers, peer)
	}
	if len(peers) == 0 {
		return nil, fmt.Errorf("no peers found")
	}
	return peers, nil
}

// Close stops underlying goroutines, if there is any. It should be a blocking call.
func (s *serviceDiscovery) Close() {}

// ServiceDiscovery defines a service discovery plugin for Olric, backed by Consul.
var ServiceDiscovery serviceDiscovery
