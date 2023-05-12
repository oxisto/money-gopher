package common

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bufbuild/connect-go"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/persistence"
	"google.golang.org/protobuf/types/known/emptypb"
)

type errorOp[T any] struct {
	delErr error
}

func (e *errorOp[T]) Replace(o persistence.StorageObject) (err error) {
	return
}

func (e *errorOp[T]) List(args ...any) (list []T, err error) {
	return
}

func (e *errorOp[T]) Get(key any) (obj T, err error) {
	return
}

func (e *errorOp[T]) Update(key any, in T, columns []string) (out T, err error) {
	return
}

func (e *errorOp[T]) Delete(key any) (err error) {
	return e.delErr
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
				op: &errorOp[*portfoliov1.Portfolio]{
					delErr: errors.New("some-error"),
				},
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
