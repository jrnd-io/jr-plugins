//go:build plugin_azblobstorage
// +build plugin_azblobstorage

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

package azblobstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/google/uuid"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
	"github.com/rs/zerolog/log"
)

const (
	Name = "azblobstorage"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	configuration Config
	client        *azblob.Client
}

func (p *Plugin) Init(ctx context.Context, cfgBytes []byte) error {
	config := Config{}
	if err := json.Unmarshal(cfgBytes, &config); err != nil {
		return err
	}

	if config.AccountName == "" {
		return fmt.Errorf("AccountName is mandatory")
	}

	if config.PrimaryAccountKey == "" {
		return fmt.Errorf("PrimaryAccountKey is mandatory")
	}

	p.configuration = config
	cred, err := azblob.NewSharedKeyCredential(config.AccountName, config.PrimaryAccountKey)
	if err != nil {
		return err
	}

	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net", config.AccountName), cred, nil)
	if err != nil {
		return err
	}

	if config.Container.Name == "" {
		return fmt.Errorf("Container name is mandatory")

	}
	if config.Container.Create {
		_, err := client.CreateContainer(ctx, config.Container.Name, nil)
		if err != nil {
			return err
		}
	}

	p.client = client
	return nil

}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string, configParams map[string]string) (*jrpc.ProduceResponse, error) {

	container := p.configuration.Container.Name
	if configParams["container.name"] != "" {
		container = configParams["container.name"]
	}

	// adding metadata to the blob
	metadata := make(map[string]*string)
	for k, v := range configParams {
		if strings.HasPrefix(k, "metadata.") {
			metadata[strings.TrimPrefix(k, "metadata.")] = &v
		}
	}
	// adding headers as metadata
	for k, v := range headers {
		metadata[k] = &v
	}

	var key string
	if len(k) == 0 || strings.ToLower(string(k)) == "null" {
		// generate a UUID as index
		key = uuid.New().String()
	} else {
		key = string(k)
	}
	metadata["key"] = &key

	resp, err := p.client.UploadBuffer(
		context.Background(),
		container,
		key,
		v,
		&azblob.UploadBufferOptions{
			Metadata: metadata,
		},
	)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("key", key).Interface("upload_resp", resp).Msg("Uploaded blob")
	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: "",
	}, nil

}

func (p *Plugin) Close(_ context.Context) error {
	return nil
}
