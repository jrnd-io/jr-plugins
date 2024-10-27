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

package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

var (
	pluginImpl Plugin
	Name       string
	lock       sync.Mutex
)

type Plugin interface {
	jrpc.Producer
	Init(context.Context, []byte) error
}

func GetPlugin() Plugin {
	return pluginImpl
}

func RegisterPlugin(name string, p Plugin) {
	lock.Lock()
	defer lock.Unlock()
	if pluginImpl != nil {
		panic(fmt.Errorf("plugin: RegisterPlugin called twice, already registered by %s", Name))
	}
	Name = name
	pluginImpl = p
}
