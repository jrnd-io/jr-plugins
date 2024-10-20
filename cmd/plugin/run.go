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
package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	hashiplugin "github.com/hashicorp/go-plugin"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/awsdynamodb"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/azblobstorage"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/azcosmosdb"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/cassandra"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/console"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/elastic"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/gcs"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/http"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/luascript"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/mongodb"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/redis"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/s3"
	_ "github.com/jrnd-io/jr-plugins/internal/plugin/wasm"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

var (
	cfgFile  string
	cfgBytes []byte
)

func init() {

	cobra.OnInitialize(readConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
}

func readConfig() {

	var err error
	if cfgFile != "" {
		cfgBytes, err = os.ReadFile(cfgFile)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read config file")

		}
	}

}

func run(_ *cobra.Command, _ []string) {
	// check registered plugin

	p := plugin.GetPlugin()
	if p == nil {
		log.Fatal().Msg("plugin instance is null")
	}

	// init plugin
	err := p.Init(context.Background(), cfgBytes)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize plugin")
	}

	hashiplugin.Serve(&hashiplugin.ServeConfig{
		HandshakeConfig: jrpc.Handshake,
		Plugins: map[string]hashiplugin.Plugin{
			"jr-plugin": &jrpc.ProducerGRPCPlugin{Impl: p},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: hashiplugin.DefaultGRPCServer,
	})

}
