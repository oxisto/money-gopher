package common

import (
	"github.com/bufbuild/connect-go"
	"github.com/oxisto/money-gopher/persistence"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Delete[T persistence.StorageObject](key any, op persistence.StorageOperations[T]) (res *connect.Response[emptypb.Empty], err error) {
	res = connect.NewResponse(&emptypb.Empty{})
	err = op.Delete(key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return
}
