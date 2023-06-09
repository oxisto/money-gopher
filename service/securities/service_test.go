// Copyright 2023 Christian Banse
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
//
// This file is part of The Money Gopher.

package securities

import (
	"testing"

	"github.com/oxisto/money-gopher/internal"

	"github.com/oxisto/assert"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name string
		want assert.Want[*service]
	}{
		{
			name: "Default",
			want: func(t *testing.T, s *service) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService(internal.NewTestDB(t))
			tt.want(t, assert.Is[*service](t, got))
		})
	}
}
