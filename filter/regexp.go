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

package filter

import (
	"github.com/trivago/gollum/core"
	"regexp"
)

// RegExp filter plugin
// This plugin allows filtering messages using regular expressions.
// Configuration example
//
//  - "stream.Broadcast":
//    Filter: "filter.RegExp"
//    FilterExpression: "\d+-.*"
//    FilterExpressionNot: "\d+-.*"
//
// FilterExpression defines the regular expression used for matching the message
// payload. If the expression matches, the message is passed.
//
// FilterExpressionNot defines a negated regular expression used for matching
// the message payload. If the expression matches, the message is blocked.
type RegExp struct {
	core.FilterBase
	exp    *regexp.Regexp
	expNot *regexp.Regexp
}

func init() {
	core.TypeRegistry.Register(RegExp{})
}

// Configure initializes this filter with values from a plugin config.
func (filter *RegExp) Configure(conf core.PluginConfigReader) error {
	var err error
	filter.FilterBase.Configure(conf)

	exp := conf.GetString("Expression", "")
	if exp != "" {
		filter.exp, err = regexp.Compile(exp)
		conf.Errors.Push(err)
	}

	exp = conf.GetString("ExpressionNot", "")
	if exp != "" {
		filter.expNot, err = regexp.Compile(exp)
		conf.Errors.Push(err)
	}

	return conf.Errors.OrNil()
}

// Accepts allows all messages
func (filter *RegExp) Accepts(msg *core.Message) bool {
	if filter.exp != nil {
		return filter.exp.MatchString(string(msg.Data))
	}

	if filter.expNot != nil {
		return !filter.expNot.MatchString(string(msg.Data))
	}

	return true // ### return, pass everything ###

}
