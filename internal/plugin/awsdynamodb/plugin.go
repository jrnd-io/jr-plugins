//go:build plugin_awsdynamodb
// +build plugin_awsdynamodb

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

package awsdynamodb

import (
	"context"
	"encoding/json"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

const (
	Name = "awsdynamodb"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	configuration Config

	client *dynamodb.Client
}

func (p *Plugin) Init(ctx context.Context, cfgBytes []byte) error {
	config := Config{}
	err := json.Unmarshal(cfgBytes, &config)
	if err != nil {
		return err
	}

	if config.Table == "" {
		return fmt.Errorf("Table is mandatory")
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := dynamodb.NewFromConfig(awsConfig)

	p.client = client
	p.configuration = config
	return nil
}

func (p *Plugin) Produce(_ []byte, val []byte, headers map[string]string, configParams map[string]string) (*jrpc.ProduceResponse, error) {

	table := p.configuration.Table
	if configParams["table"] != "" {
		table = configParams["table"]

	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(val, &jsonMap); err != nil {
		return nil, err
	}

	item, err := attributevalue.MarshalMap(jsonMap)
	if err != nil {
		return nil, err
	}

	_, err = p.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	if err != nil {
		return nil, err
	}

	return &jrpc.ProduceResponse{
		Bytes: uint64(len(val)),
	}, nil

}

func (p *Plugin) Close(_ context.Context) error {
	return nil
}
