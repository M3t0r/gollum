// Copyright 2015-2016 trivago GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package format

import (
	"github.com/trivago/gollum/core"
)

// Clear formatter plugin
// Clear is a formatter that clears the message
// Configuration example
//
//  - "stream.Broadcast":
//    Formatter: "format.Clear"
type Clear struct {
}

func init() {
	core.TypeRegistry.Register(Clear{})
}

// Configure initializes this formatter with values from a plugin config.
func (format *Clear) Configure(conf core.PluginConfigReader) error {
	return nil
}

// Modulate replaces the current payload by an empty buffer
func (format *Clear) Modulate(msg *core.Message) core.ModulateResult {
	msg.Store([]byte{})
	return core.ModulateResultContinue
}
