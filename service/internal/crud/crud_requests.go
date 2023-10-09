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

// package crud contains helpers to handle CRUD (Create, Read, Update and
// Delete) requests that work on [persistence.StorageOperations] in a common
// way.
package crud

import (
	"github.com/oxisto/money-gopher/persistence"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Create[T any, S persistence.StorageObject](obj S, op persistence.StorageOperations[S], convert func(obj S) *T) (res *connect.Response[T], err error) {
	// TODO(oxisto): We probably want to have a pure create instead of replace here
	err = op.Replace(obj)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res = connect.NewResponse(convert(obj))

	return
}

func List[T any, S persistence.StorageObject](op persistence.StorageOperations[S], setter func(res *connect.Response[T], list []S), args ...any) (res *connect.Response[T], err error) {
	obj, err := op.List(args...)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res = connect.NewResponse(new(T))
	setter(res, obj)

	return
}

func Get[T any, S persistence.StorageObject](key any, op persistence.StorageOperations[S], convert func(obj S) *T) (res *connect.Response[T], err error) {
	obj, err := op.Get(key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res = connect.NewResponse(convert(obj))

	return
}

func Update[T any, S persistence.StorageObject](key any, in S, paths []string, op persistence.StorageOperations[S], convert func(obj S) *T) (res *connect.Response[T], err error) {
	out, err := op.Update(key, in, paths)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res = connect.NewResponse(convert(out))

	return
}

func Delete[S persistence.StorageObject](key any, op persistence.StorageOperations[S]) (res *connect.Response[emptypb.Empty], err error) {
	err = op.Delete(key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res = connect.NewResponse(&emptypb.Empty{})

	return
}
