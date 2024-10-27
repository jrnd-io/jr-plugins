//go:build plugin_luascript
// +build plugin_luascript

// Copyright © 2024 JR team
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

package luascript_test

import (
	"testing"

	"github.com/jrnd-io/jr-plugins/internal/plugin/luascript"
)

func TestProducer(t *testing.T) {

	testCases := []struct {
		name   string
		config luascript.Config
	}{
		{
			name: "testprint",
			config: luascript.Config{
				Script: ` json = require "json"

				          print(k)
						  json.decode(v)
						  if headers["h1"] ~= "v1" then
						     error("header mismatch")
					      end
						  if headers["h2"] ~= "v2" then
						     error("header mismatch")
					      end
						  `,
			},
		},
	}
	for _, tc := range testCases {
		someJSON := `{"key": "value"}`

		t.Run(tc.name, func(_ *testing.T) {
			p := &luascript.Plugin{}
			err := p.InitFromConfig(tc.config)
			if err != nil {
				t.Error(err)
			}
			_, err = p.Produce([]byte("somekey"),
				[]byte(someJSON),
				map[string]string{
					"h1": "v1",
					"h2": "v2",
				}, nil)
			if err != nil {
				t.Error(err)
			}
		})

	}
}
