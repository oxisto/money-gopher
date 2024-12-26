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
	"context"
	"slices"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/service/internal/crud"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (svc *service) CreateSecurity(ctx context.Context, req *connect.Request[portfoliov1.CreateSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	return crud.Create(
		req.Msg.Security,
		svc.securities,
		func(obj *portfoliov1.Security) *portfoliov1.Security {
			for _, ls := range obj.ListedOn {
				svc.listedSecurities.Replace(ls)
			}

			return obj
		},
	)
}

func (svc *service) GetSecurity(ctx context.Context, req *connect.Request[portfoliov1.GetSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	return crud.Get(
		req.Msg.Name,
		svc.securities,
		func(obj *portfoliov1.Security) *portfoliov1.Security {
			obj.ListedOn, _ = svc.listedSecurities.List(obj.Id)

			return obj
		},
	)
}

func (svc *service) ListSecurities(ctx context.Context, req *connect.Request[portfoliov1.ListSecuritiesRequest]) (res *connect.Response[portfoliov1.ListSecuritiesResponse], err error) {
	return crud.List(
		svc.securities,
		func(res *connect.Response[portfoliov1.ListSecuritiesResponse], list []*portfoliov1.Security) error {
			res.Msg.Securities = list

			for _, sec := range res.Msg.Securities {
				sec.ListedOn, err = svc.listedSecurities.List(sec.Id)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
}

func (svc *service) UpdateSecurity(ctx context.Context, req *connect.Request[portfoliov1.UpdateSecurityRequest]) (res *connect.Response[portfoliov1.Security], err error) {
	return crud.Update(
		req.Msg.Security.Id,
		req.Msg.Security,
		req.Msg.UpdateMask.Paths,
		svc.securities,
		func(obj *portfoliov1.Security) *portfoliov1.Security {
			if slices.Contains(req.Msg.UpdateMask.Paths, "listed_on") {
				for _, ls := range req.Msg.Security.ListedOn {
					svc.listedSecurities.Replace(ls)
				}
			}

			return obj
		},
	)
}

func (svc *service) DeleteSecurity(ctx context.Context, req *connect.Request[portfoliov1.DeleteSecurityRequest]) (res *connect.Response[emptypb.Empty], err error) {
	return crud.Delete(
		req.Msg.Name,
		svc.securities,
	)
}

func (svc *service) fetchSecurity(name string) (sec *portfoliov1.Security, err error) {
	res, err := crud.Get(
		name,
		svc.securities,
		func(obj *portfoliov1.Security) *portfoliov1.Security {
			obj.ListedOn, _ = svc.listedSecurities.List(obj.Id)

			return obj
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Msg, nil
}
