//go:build plugin_redis
// +build plugin_redis

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

package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	Name = "redis"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	client redis.Client
	Ttl    time.Duration
}

func (p *Plugin) Init(_ context.Context, cfgBytes []byte) error {
	var options redis.Options

	err := json.Unmarshal(cfgBytes, &options)
	if err != nil {
		return err
	}

	p.client = *redis.NewClient(&options)
	return nil
}

func (p *Plugin) Close(_ context.Context) error {
	err := p.client.Close()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to close Redis connection")
	}
	return err
}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string, _ map[string]string) (*jrpc.ProduceResponse, error) {
	err := p.client.Set(context.Background(), string(k), string(v), p.Ttl).Err()
	if err != nil {
		return nil, err
	}
	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: "",
	}, nil
}
