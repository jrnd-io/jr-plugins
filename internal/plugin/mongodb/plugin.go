//go:build plugin_mongodb
// +build plugin_mongodb

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

package mongodb

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

type Config struct {
	MongoURI   string `json:"mongo_uri"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

const (
	Name = "mongodb"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	client     mongo.Client
	database   string
	collection string
}

func (p *Plugin) Init(ctx context.Context, configBytes []byte) error {
	config := Config{}
	err := json.Unmarshal(configBytes, &config)
	if err != nil {
		return err
	}

	clientOptions := options.Client().ApplyURI(config.MongoURI)
	if config.Username != "" && config.Password != "" {
		clientOptions.Auth = &options.Credential{
			Username: config.Username,
			Password: config.Password,
		}
	}

	p.collection = config.Collection
	p.database = config.Database

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return err
	}

	p.client = *client
	return nil
}

func (p *Plugin) Produce(key []byte, v []byte, headers map[string]string, configParams map[string]string) (*jrpc.ProduceResponse, error) {

	database := p.database
	if configParams["database"] != "" {
		database = configParams["database"]
	}
	collectionName := p.collection
	if configParams["collection"] != "" {
		collectionName = configParams["collection"]
	}

	collection := p.client.Database(database).Collection(collectionName)

	var dev map[string]interface{}
	err := json.Unmarshal(v, &dev)
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		dev["_id"] = string(key)
	}

	resp, err := collection.InsertOne(context.Background(), dev)
	if err != nil {
		return nil, err
	}

	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: resp.InsertedID.(string),
	}, nil
}

func (p *Plugin) Close(ctx context.Context) error {
	err := p.client.Disconnect(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to close Mongo connection")
	}
	return err
}
