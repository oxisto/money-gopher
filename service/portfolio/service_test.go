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

package portfolio

import (
	"testing"

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
)

func TestNewService(t *testing.T) {
	type args struct {
		opts Options
	}
	tests := []struct {
		name string
		args args
		want assert.Want[portfoliov1connect.PortfolioServiceHandler]
	}{
		{
			name: "with default client",
			args: args{opts: Options{}},
			want: func(t *testing.T, psh portfoliov1connect.PortfolioServiceHandler) bool {
				s := assert.Is[*service](t, psh)
				return assert.NotNil(t, s.securities)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService(tt.args.opts)
			tt.want(t, got)
		})
	}
}
