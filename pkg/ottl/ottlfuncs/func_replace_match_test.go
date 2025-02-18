// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ottlfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

func Test_replaceMatch(t *testing.T) {
	input := pcommon.NewValueStr("hello world")

	target := &ottl.StandardGetSetter[pcommon.Value]{
		Getter: func(ctx pcommon.Value) (interface{}, error) {
			return ctx.Str(), nil
		},
		Setter: func(ctx pcommon.Value, val interface{}) error {
			ctx.SetStr(val.(string))
			return nil
		},
	}

	tests := []struct {
		name        string
		target      ottl.GetSetter[pcommon.Value]
		pattern     string
		replacement string
		want        func(pcommon.Value)
	}{
		{
			name:        "replace match",
			target:      target,
			pattern:     "hello*",
			replacement: "hello {universe}",
			want: func(expectedValue pcommon.Value) {
				expectedValue.SetStr("hello {universe}")
			},
		},
		{
			name:        "no match",
			target:      target,
			pattern:     "goodbye*",
			replacement: "goodbye {universe}",
			want: func(expectedValue pcommon.Value) {
				expectedValue.SetStr("hello world")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenarioValue := pcommon.NewValueStr(input.Str())

			exprFunc, err := ReplaceMatch(tt.target, tt.pattern, tt.replacement)
			assert.NoError(t, err)
			result, err := exprFunc(scenarioValue)
			assert.NoError(t, err)
			assert.Nil(t, result)

			expected := pcommon.NewValueStr("")
			tt.want(expected)

			assert.Equal(t, expected, scenarioValue)
		})
	}
}

func Test_replaceMatch_bad_input(t *testing.T) {
	input := pcommon.NewValueInt(1)
	target := &ottl.StandardGetSetter[interface{}]{
		Getter: func(ctx interface{}) (interface{}, error) {
			return ctx, nil
		},
		Setter: func(ctx interface{}, val interface{}) error {
			t.Errorf("nothing should be set in this scenario")
			return nil
		},
	}

	exprFunc, err := ReplaceMatch[interface{}](target, "*", "{replacement}")
	assert.NoError(t, err)

	result, err := exprFunc(input)
	assert.NoError(t, err)
	assert.Nil(t, result)

	assert.Equal(t, pcommon.NewValueInt(1), input)
}

func Test_replaceMatch_get_nil(t *testing.T) {
	target := &ottl.StandardGetSetter[interface{}]{
		Getter: func(ctx interface{}) (interface{}, error) {
			return ctx, nil
		},
		Setter: func(ctx interface{}, val interface{}) error {
			t.Errorf("nothing should be set in this scenario")
			return nil
		},
	}

	exprFunc, err := ReplaceMatch[interface{}](target, "*", "{anything}")
	assert.NoError(t, err)

	result, err := exprFunc(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
