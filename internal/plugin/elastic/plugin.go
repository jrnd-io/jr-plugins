//go:build plugin_elastic
// +build plugin_elastic

// Copyright © 2024 JR team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package elastic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

type Config struct {
	ElasticURI      string `json:"es_uri"`
	ElasticIndex    string `json:"index"`
	ElasticUsername string `json:"username"`
	ElasticPassword string `json:"password"`
}

const (
	Name = "elastic"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	client *elasticsearch.Client
	index  string
}

func (p *Plugin) Init(_ context.Context, cfgBytes []byte) error {
	config := Config{}
	err := json.Unmarshal(cfgBytes, &config)
	if err != nil {
		return err
	}

	cfg := elasticsearch.Config{
		Addresses: []string{config.ElasticURI},
		Username:  config.ElasticUsername,
		Password:  config.ElasticPassword,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	p.index = config.ElasticIndex
	p.client = client
	return nil
}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string, _ map[string]string) (*jrpc.ProduceResponse, error) {

	var req esapi.IndexRequest

	if len(k) == 0 {
		// generate a UUID as index
		id := uuid.New()

		req = esapi.IndexRequest{
			Index:      p.index,
			DocumentID: id.String(),
			Body:       strings.NewReader(string(v)),
			Refresh:    "true",
		}
	} else {
		req = esapi.IndexRequest{
			Index:      p.index,
			DocumentID: string(k),
			Body:       strings.NewReader(string(v)),
			Refresh:    "true",
		}
	}

	res, err := req.Do(context.Background(), p.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("error: %s", body)
	}

	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: string(body),
	}, nil
}

func (p *Plugin) Close(_ context.Context) error {
	return nil
}
