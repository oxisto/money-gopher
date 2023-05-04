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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oxisto/assert"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

type mockRoundTripper struct {
	f func(req *http.Request) (res *http.Response, err error)
}

func (t *mockRoundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	return t.f(req)
}

func newMockClient(f func(req *http.Request) (res *http.Response, err error)) (c http.Client) {
	return http.Client{
		Transport: &mockRoundTripper{f},
	}
}

func Test_yf_LatestQuote(t *testing.T) {
	type fields struct {
		Client http.Client
	}
	type args struct {
		ls *portfoliov1.ListedSecurity
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantQuote float32
		wantTime  time.Time
		wantErr   bool
	}{
		{
			name: "happy path",
			fields: fields{
				Client: newMockClient(func(req *http.Request) (res *http.Response, err error) {
					r := httptest.NewRecorder()
					r.WriteString(`{"chart":{"result":[{"meta": {"regularMarketPrice": 100.0, "regularMarketTime": 1683230400}}]}}`)
					return r.Result(), nil
				}),
			},
			args: args{
				ls: &portfoliov1.ListedSecurity{
					SecurityName: "My Security",
					Ticker:       "TICK",
					Currency:     "USD",
				},
			},
			wantQuote: float32(100.0),
			wantTime:  time.Date(2023, 05, 04, 22, 0, 0, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yf := &yf{
				Client: tt.fields.Client,
			}
			gotQuote, gotTime, err := yf.LatestQuote(tt.args.ls)
			if (err != nil) != tt.wantErr {
				t.Errorf("yf.LatestQuote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equals(t, tt.wantQuote, gotQuote)
			assert.Equals(t, tt.wantTime, gotTime)
		})
	}
}
