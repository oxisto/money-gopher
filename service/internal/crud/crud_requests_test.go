package crud

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/oxisto/assert"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/persistence"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestCreate(t *testing.T) {
	type args struct {
		obj     *portfoliov1.Portfolio
		op      persistence.StorageOperations[*portfoliov1.Portfolio]
		convert func(obj *portfoliov1.Portfolio) *portfoliov1.Portfolio
	}
	tests := []struct {
		name    string
		args    args
		wantRes *connect.Response[portfoliov1.Portfolio]
		wantErr bool
	}{
		{
			name: "error",
			args: args{
				op: internal.ErrOps[*portfoliov1.Portfolio](errors.New("some-error")),
				convert: func(obj *portfoliov1.Portfolio) *portfoliov1.Portfolio {
					return obj
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := Create(tt.args.obj, tt.args.op, tt.args.convert)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("Create() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestList(t *testing.T) {
	type args struct {
		op     persistence.StorageOperations[*portfoliov1.Portfolio]
		setter func(res *connect.Response[portfoliov1.ListPortfoliosResponse], list []*portfoliov1.Portfolio)
		args   []any
	}
	tests := []struct {
		name    string
		args    args
		wantRes *connect.Response[portfoliov1.ListPortfoliosResponse]
		wantErr bool
	}{
		{
			name: "error",
			args: args{
				op: internal.ErrOps[*portfoliov1.Portfolio](errors.New("some-error")),
				setter: func(res *connect.Response[portfoliov1.ListPortfoliosResponse], list []*portfoliov1.Portfolio) {
					res.Msg.Portfolios = list
				},
				args: []any{"some-key"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := List(tt.args.op, tt.args.setter, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equals(t, tt.wantRes, gotRes) {
				t.Errorf("List() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		key     any
		op      persistence.StorageOperations[*portfoliov1.Portfolio]
		convert func(obj *portfoliov1.Portfolio) *portfoliov1.Portfolio
	}
	tests := []struct {
		name    string
		args    args
		wantRes *connect.Response[portfoliov1.Portfolio]
		wantErr bool
	}{
		{
			name: "error",
			args: args{
				key: "some-key",
				op:  internal.ErrOps[*portfoliov1.Portfolio](errors.New("some-error")),
				convert: func(obj *portfoliov1.Portfolio) *portfoliov1.Portfolio {
					return obj
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := Get(tt.args.key, tt.args.op, tt.args.convert)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("Get() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		key     any
		in      *portfoliov1.Portfolio
		paths   []string
		op      persistence.StorageOperations[*portfoliov1.Portfolio]
		convert func(obj *portfoliov1.Portfolio) *portfoliov1.Portfolio
	}
	tests := []struct {
		name    string
		args    args
		wantRes *connect.Response[portfoliov1.Portfolio]
		wantErr bool
	}{
		{
			name: "error",
			args: args{
				key: "some-key",
				op:  internal.ErrOps[*portfoliov1.Portfolio](errors.New("some-error")),
				convert: func(obj *portfoliov1.Portfolio) *portfoliov1.Portfolio {
					return obj
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := Update(tt.args.key, tt.args.in, tt.args.paths, tt.args.op, tt.args.convert)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("Update() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		key any
		op  persistence.StorageOperations[*portfoliov1.Portfolio]
	}
	tests := []struct {
		name    string
		args    args
		wantRes *connect.Response[emptypb.Empty]
		wantErr bool
	}{
		{
			name: "error",
			args: args{
				key: "some-key",
				op:  internal.ErrOps[*portfoliov1.Portfolio](errors.New("some-error")),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := Delete(tt.args.key, tt.args.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("Delete() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
