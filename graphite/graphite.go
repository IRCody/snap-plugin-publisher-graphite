/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package graphite

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	plh "github.com/intelsdi-x/snap-plugin-publisher-graphite/logHelper"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/marpaia/graphite-golang"
)

const (
	name       = "graphite"
	version    = 1
	pluginType = plugin.PublisherPluginType
)

type graphitePublisher struct {
}

func NewGraphitePublisher() *graphitePublisher {
	return &graphitePublisher{}
}

func (f *graphitePublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {

	logger := plh.GetLogger(config, Meta())
	logger.Debug("Publishing started")
	var metrics []plugin.PluginMetricType

	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			logger.Errorf("Error decoding: error=%v content=%v", err, content)
			return err
		}
	default:
		logger.Errorf("Error unknown content type '%v'", contentType)
		return fmt.Errorf("Unknown content type '%s'", contentType)
	}

	logger.Debug("publishing %v metrics to %v", len(metrics), config)

	server := config["server"].(ctypes.ConfigValueStr).Value
	port := config["port"].(ctypes.ConfigValueInt).Value

	logger.Debug("Attempting to connect to %s:%d", server, port)
	gite, err := graphite.NewGraphite(server, port)
	if err != nil {
		logger.Errorf("Error Connecting to graphite at %s:%d. Error: %v", server, port, err)
		return fmt.Errorf("Error Connecting to graphite at %s:%d. Error: %v", server, port, err)
	}
	logger.Debug("Connected to %s:%s successfully", server, port)
	for _, m := range metrics {
		key := strings.Join(m.Namespace(), ".")
		data := fmt.Sprintf("%v", m.Data())
		logger.Debug("Attempting to send %s:%s", key, data)
		err = gite.SimpleSend(key, data)
		if err != nil {
			logger.Errorf("Unable to send metric %s:%s to %s:%d. Error: %s", key, data, server, port, err)
			return fmt.Errorf("Unable to send metric %s:%s to %s:%d. Error: %s", key, data, server, port, err)
		}
		logger.Debug("Sent %s, %s", key, data)
	}
	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

func (f *graphitePublisher) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()
	config, err := plh.AddLogging(config)
	if err != nil {
		config = cpolicy.NewPolicyNode()
	}
	r1, err := cpolicy.NewStringRule("server", true)
	handleErr(err)
	r1.Description = "Address of graphite server"
	config.Add(r1)

	r2, err := cpolicy.NewIntegerRule("port", true)
	handleErr(err)
	r2.Description = "Port to connect on"
	config.Add(r2)

	cp.Add([]string{""}, config)
	fmt.Println(config)
	return cp, nil
}

func handleErr(e error) {
	if e != nil {
		log.Panic(e)
		panic(e)
	}
}
