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
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/service/internal/crud"
)

var bankAccountSetter = func(obj *portfoliov1.BankAccount) *portfoliov1.BankAccount {
	return obj
}

func (svc *service) CreateBankAccount(ctx context.Context, req *connect.Request[portfoliov1.CreateBankAccountRequest]) (res *connect.Response[portfoliov1.BankAccount], err error) {
	return crud.Create(
		req.Msg.BankAccount,
		svc.bankAccounts,
		bankAccountSetter,
	)
}

func (svc *service) UpdateBankAccount(ctx context.Context, req *connect.Request[portfoliov1.UpdateBankAccountRequest]) (res *connect.Response[portfoliov1.BankAccount], err error) {
	return crud.Update(
		req.Msg.Account.Name,
		req.Msg.Account,
		req.Msg.UpdateMask.Paths,
		svc.bankAccounts,
		func(obj *portfoliov1.BankAccount) *portfoliov1.BankAccount {
			return obj
		},
	)
}

func (svc *service) DeleteBankAccount(ctx context.Context, req *connect.Request[portfoliov1.DeleteBankAccountRequest]) (res *connect.Response[emptypb.Empty], err error) {
	return crud.Delete(req.Msg.Name, svc.bankAccounts)
}
