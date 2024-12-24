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

package commands

import (
	"testing"
	"time"

	"github.com/oxisto/assert"
	"github.com/oxisto/money-gopher/cli"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/testing/servertest"
	"github.com/posener/complete"
)

func TestCreatePortfolioCmd_Run(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type fields struct {
		Name        string
		DisplayName string
	}
	type args struct {
		s *cli.Session
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Name:        "myportfolio",
				DisplayName: "My Portfolio",
			},
			args: args{
				s: func() *cli.Session {
					return cli.NewSession(&cli.SessionOptions{
						BaseURL:    srv.URL,
						HttpClient: srv.Client(),
					})
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CreatePortfolioCmd{
				Name:        tt.fields.Name,
				DisplayName: tt.fields.DisplayName,
			}
			if err := cmd.Run(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("CreatePortfolioCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShowPortfolioCmd_Run(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type fields struct {
		PortfolioName string
	}
	type args struct {
		s *cli.Session
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				PortfolioName: "myportfolio",
			},
			args: args{
				s: func() *cli.Session {
					return cli.NewSession(&cli.SessionOptions{
						BaseURL:    srv.URL,
						HttpClient: srv.Client(),
					})
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ShowPortfolioCmd{
				PortfolioName: tt.fields.PortfolioName,
			}
			if err := cmd.Run(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ShowPortfolioCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateTransactionCmd_Run(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type fields struct {
		PortfolioName string
		SecurityId    string
		Type          string
		Amount        float64
		Price         float32
		Fees          float32
		Taxes         float32
		Time          time.Time
	}
	type args struct {
		s *cli.Session
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				PortfolioName: "myportfolio",
				Price:         10.0,
			},
			args: args{
				s: func() *cli.Session {
					return cli.NewSession(&cli.SessionOptions{
						BaseURL:    srv.URL,
						HttpClient: srv.Client(),
					})
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CreateTransactionCmd{
				PortfolioName: tt.fields.PortfolioName,
				SecurityId:    tt.fields.SecurityId,
				Type:          tt.fields.Type,
				Amount:        tt.fields.Amount,
				Price:         tt.fields.Price,
				Fees:          tt.fields.Fees,
				Taxes:         tt.fields.Taxes,
				Time:          tt.fields.Time,
			}
			if err := cmd.Run(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("CreateTransactionCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListPortfolioCmd_Run(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type args struct {
		s *cli.Session
	}
	tests := []struct {
		name    string
		l       *ListPortfolioCmd
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				s: func() *cli.Session {
					return cli.NewSession(&cli.SessionOptions{
						BaseURL:    srv.URL,
						HttpClient: srv.Client(),
					})
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &ListPortfolioCmd{}
			if err := l.Run(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ListPortfolioCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPredictPortfolios(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type args struct {
		s *cli.Session
		a complete.Args
	}
	tests := []struct {
		name string
		args args
		want assert.Want[[]string]
	}{
		{
			name: "happy path",
			args: args{
				s: func() *cli.Session {
					return cli.NewSession(&cli.SessionOptions{
						BaseURL:    srv.URL,
						HttpClient: srv.Client(),
					})
				}(),
				a: complete.Args{
					All:  []string{},
					Last: "my",
				},
			},
			want: func(t *testing.T, s []string) bool {
				return len(s) > 0
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := PredictPortfolios(tt.args.s)
			got := fn.Predict(tt.args.a)
			tt.want(t, got)
		})
	}
}
