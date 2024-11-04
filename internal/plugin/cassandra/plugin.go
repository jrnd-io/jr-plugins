//go:build plugin_cassandra
// +build plugin_cassandra

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

package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
	"github.com/rs/zerolog/log"
)

const (
	Name = "cassandra"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	configuration Config

	session          *gocql.Session
	consistencyLevel gocql.Consistency
	timeout          time.Duration
}

func (p *Plugin) Init(_ context.Context, cfgBytes []byte) error {
	config := Config{}
	if err := json.Unmarshal(cfgBytes, &config); err != nil {
		return err
	}

	if config.Keyspace == "" {
		return fmt.Errorf("Keyspace is required")
	}

	if len(config.Hosts) == 0 {
		return fmt.Errorf("Hosts are required")
	}

	if config.Username == "" || config.Password == "" {
		return fmt.Errorf("Username and password are both required")
	}

	if config.ConsistencyLevel == "" {
		config.ConsistencyLevel = "QUORUM"
	}

	consistencyLevel, err := gocql.MustParseConsistency(config.ConsistencyLevel)
	if err != nil {
		return err
	}

	if config.Timeout == "" {
		config.Timeout = "10s"
	}

	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse timeout, setting default to 10s")
		timeout = time.Second * 10
	}

	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Username,
		Password: config.Password,
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	p.configuration = config
	p.timeout = timeout
	p.session = session
	p.consistencyLevel = consistencyLevel

	return nil

}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string, configParams map[string]string) (*jrpc.ProduceResponse, error) {

	keyspace := p.configuration.Keyspace
	if configParams["keyspace"] != "" {
		keyspace = configParams["keyspace"]
	}
	table := p.configuration.Table
	if configParams["table"] != "" {
		table = configParams["table"]
	}
	consistencyLevel := p.consistencyLevel
	if configParams["consistency_level"] != "" {
		consistencyLevel, err := gocql.MustParseConsistency(configParams["consistency_level"])
		if err != nil {
			return nil, err
		}
		p.consistencyLevel = consistencyLevel
	}

	stmt := fmt.Sprintf("INSERT INTO %s.%s JSON ?", keyspace, table)
	if err := p.session.Query(stmt, string(v)).
		Consistency(consistencyLevel).Exec(); err != nil {
		return nil, err
	}
	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: "",
	}, nil
}

func (p *Plugin) Close(_ context.Context) error {
	p.session.Close()
	return nil
}
