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

package quote

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/persistence"

	"github.com/oxisto/assert"
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
		ctx context.Context
		ls  *persistence.ListedSecurity
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantQuote *currency.Currency
		wantTime  time.Time
		wantErr   assert.Want[error]
	}{
		{
			name: "http response error",
			fields: fields{
				Client: newMockClient(func(req *http.Request) (res *http.Response, err error) {
					return nil, http.ErrNotSupported
				}),
			},
			args: args{
				ctx: context.TODO(),
				ls: &persistence.ListedSecurity{
					SecurityID: "My Security",
					Ticker:     "TICK",
					Currency:   "USD",
				},
			},
			wantErr: func(t *testing.T, err error) bool {
				return errors.Is(err, http.ErrNotSupported)
			},
		},
		{
			name: "invalid JSON",
			fields: fields{
				Client: newMockClient(func(req *http.Request) (res *http.Response, err error) {
					r := httptest.NewRecorder()
					r.WriteString(`{]`)
					return r.Result(), nil
				}),
			},
			args: args{
				ls: &persistence.ListedSecurity{
					SecurityID: "My Security",
					Ticker:     "TICK",
					Currency:   "USD",
				},
			},
			wantErr: func(t *testing.T, err error) bool {
				return strings.Contains(err.Error(), "invalid")
			},
		},
		{
			name: "invalid JSON",
			fields: fields{
				Client: newMockClient(func(req *http.Request) (res *http.Response, err error) {
					r := httptest.NewRecorder()
					r.WriteString(`{"chart":{"result":[]}}`)
					return r.Result(), nil
				}),
			},
			args: args{
				ls: &persistence.ListedSecurity{
					SecurityID: "My Security",
					Ticker:     "TICK",
					Currency:   "USD",
				},
			},
			wantErr: func(t *testing.T, err error) bool {
				return errors.Is(err, ErrEmptyResult)
			},
		},
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
				ls: &persistence.ListedSecurity{
					SecurityID: "My Security",
					Ticker:     "TICK",
					Currency:   "USD",
				},
			},
			wantQuote: currency.Value(10000),
			wantTime:  time.Date(2023, 05, 04, 20, 0, 0, 0, time.UTC),
			wantErr:   func(t *testing.T, err error) bool { return true },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yf := &yf{
				Client: tt.fields.Client,
			}
			gotQuote, gotTime, err := yf.LatestQuote(tt.args.ctx, tt.args.ls)
			assert.Equals(t, true, tt.wantErr(t, err))
			assert.Equals(t, tt.wantQuote, gotQuote)
			assert.Equals(t, tt.wantTime.UTC(), gotTime.UTC())
		})
	}
}