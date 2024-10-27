//go:build plugin_console
// +build plugin_console

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

package console

import (
	"context"
	"fmt"

	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

const (
	Name = "console"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct{}

func (p *Plugin) Init(_ context.Context, _ []byte) error {
	// No initialization needed for console plugin
	return nil
}

func (p *Plugin) Close(_ context.Context) error {
	// No cleanup needed for console plugin
	return nil
}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string, configParams map[string]string) (*jrpc.ProduceResponse, error) {
	fmt.Printf("Key: %s\n", string(k))
	fmt.Printf("Value: %s\n", string(v))
	fmt.Println("Headers:")
	for key, value := range headers {
		fmt.Printf("  %s: %s\n", key, value)
	}
	fmt.Println("ConfigParameters:")
	for key, value := range configParams {
		fmt.Printf("  %s: %s\n", key, value)
	}

	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: "Printed to console",
	}, nil
}
