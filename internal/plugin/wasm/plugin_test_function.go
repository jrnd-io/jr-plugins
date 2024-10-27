//go:build tinygo.wasm

// tinygo build -o internal/plugin/wasm/plugin_test_function.wasm -target=wasi internal/plugin/wasm/plugin_test_function.go

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
	"bytes"
	"encoding/json"
	"io"
	"os"
)

const NoError = 0

type record struct {
	K []byte            `json:"k,omitempty"`
	V []byte            `json:"v,omitempty"`
	H map[string]string `json:"h,omitempty"`
}

//export produce
func _produce(size uint32) uint64 {
	b := make([]byte, size)

	_, err := io.ReadAtLeast(os.Stdin, b, int(size))
	if err != nil {
		return e(err)
	}

	in := record{}

	err = json.Unmarshal(b, &in)
	if err != nil {
		return e(err)
	}

	out := bytes.ToUpper(in.V)

	_, err = os.Stdout.Write(out)
	if err != nil {
		return e(err)
	}

	return NoError
}

func e(err error) uint64 {
	if err == nil {
		return NoError
	}

	_, _ = os.Stderr.WriteString(err.Error())
	return uint64(len(err.Error()))
}

func main() {}
