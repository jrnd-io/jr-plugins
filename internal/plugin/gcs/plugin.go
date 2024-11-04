//go:build plugin_gcs
// +build plugin_gcs

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

package gcs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

const (
	Name = "gcs"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Config struct {
	Bucket string `json:"bucket_name"`
}

type Plugin struct {
	client storage.Client
	bucket string
}

func (p *Plugin) Init(ctx context.Context, cfgBytes []byte) error {
	config := Config{}
	err := json.Unmarshal(cfgBytes, &config)
	if err != nil {
		return err
	}

	// Use Google Application Default Credentials to authorize and authenticate the client.
	// More information about Application Default Credentials and how to enable is at
	// https://developers.google.com/identity/protocols/application-default-credentials.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	p.client = *client
	p.bucket = config.Bucket
	return nil
}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string, configParams map[string]string) (*jrpc.ProduceResponse, error) {
	bucket := p.bucket
	if configParams["bucket"] != "" {
		bucket = configParams["bucket"]
	}

	var key string

	if len(k) == 0 || strings.ToLower(string(k)) == "null" {
		// generate a UUID as index
		key = uuid.New().String()
	} else {
		key = string(k)
	}

	objectHandle := p.client.Bucket(bucket).Object(key)
	writer := objectHandle.NewWriter(context.Background())
	kvPair := fmt.Sprintf("%s=%s\n", key, v)

	b, err := writer.Write([]byte(kvPair))
	if err != nil {
		return nil, err
	}

	writer.Close()

	return &jrpc.ProduceResponse{
		Bytes:   uint64(b),
		Message: "",
	}, nil

}

func (p *Plugin) Close(_ context.Context) error {
	p.client.Close()
	return nil
}
