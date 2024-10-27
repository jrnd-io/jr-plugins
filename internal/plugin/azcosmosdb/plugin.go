//go:build plugin_azcosmosdb
// +build plugin_azcosmosdb

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

package azcosmosdb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
	"github.com/rs/zerolog/log"
)

const (
	Name = "azcosmosdb"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	configuration Config
	client        *azcosmos.Client
}

func (p *Plugin) Init(_ context.Context, cfgBytes []byte) error {
	config := Config{}
	if err := json.Unmarshal(cfgBytes, &config); err != nil {
		return err
	}

	if config.Endpoint == "" {
		return fmt.Errorf("Endpoint is mandatory")
	}

	if config.PrimaryAccountKey == "" {
		return fmt.Errorf("PrimaryAccountKey is mandatory")
	}

	if config.PartitionKey == "" {
		return fmt.Errorf("PartitionKey is mandatory")
	}

	cred, err := azcosmos.NewKeyCredential(config.PrimaryAccountKey)
	if err != nil {
		return err
	}

	client, err := azcosmos.NewClientWithKey(config.Endpoint, cred, nil)
	if err != nil {
		return err
	}

	p.configuration = config
	p.client = client
	return nil

}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string, _ map[string]string) (*jrpc.ProduceResponse, error) {

	// This is ugly but it works
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(v, &jsonMap); err != nil {
		return nil, err
	}

	// getting partition key value
	pkValue := jsonMap[p.configuration.PartitionKey]
	if pkValue == nil {
		return nil, fmt.Errorf("Partition key not found in value")
	}
	log.Debug().Str("pkValue", pkValue.(string)).Msg("Partition key value")

	container, err := p.client.NewContainer(p.configuration.Database, p.configuration.Container)
	if err != nil {
		return nil, err
		log.Fatal().Err(err).Msg("Failed to create container")
	}

	pk := azcosmos.NewPartitionKeyString(pkValue.(string))
	resp, err := container.CreateItem(context.Background(), pk, v, nil)
	if err != nil {
		return nil, err
		log.Fatal().Err(err).Msg("Failed to create item")
	}

	log.Debug().Interface("resp", resp).Msg("Item created")
	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: string(resp.ETag),
	}, nil

}

func (p *Plugin) Close(_ context.Context) error {
	return nil
}
