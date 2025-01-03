//go:build plugin_wasm
// +build plugin_wasm

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

package wasm_test

import (
	"context"
	"testing"

	"github.com/jrnd-io/jr-plugins/internal/plugin/wasm"
	"github.com/stretchr/testify/assert"
)

func TestWASMPlugin(t *testing.T) {

	testCases := []struct {
		name   string
		config wasm.Config
	}{
		{
			name: "testprint",
			config: wasm.Config{
				ModulePath: "plugin_test_function.wasm",
				BindStdout: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &wasm.Plugin{}

			ctx := context.Background()

			err := p.InitializeFromConfig(ctx, tc.config)
			assert.NoError(t, err)

			_, err = p.Produce([]byte("somekey"), []byte("someval"), nil, nil)
			assert.NoError(t, err)
		})

	}
}
