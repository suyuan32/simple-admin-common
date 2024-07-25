// Copyright 2023 The Ryan SU Authors (https://github.com/suyuan32). All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deptctx

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

// GetDepartmentIDFromCtx returns department id from context.
func GetDepartmentIDFromCtx(ctx context.Context) (uint64, error) {
	var departmentId string

	if deptId, ok := ctx.Value("deptId").(json.Number); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get department id from context", logx.Field("detail", ctx))
			return 0, errorx.NewInvalidArgumentError("failed to get department ID")
		} else {
			if data := md.Get(enum.DepartmentIdRpcCtxKey); len(data) > 0 {
				departmentId = data[0]
			} else {
				return 0, errorx.NewInvalidArgumentError("failed to get department ID")
			}
		}
	} else {
		departmentId = deptId.String()
	}

	id, err := strconv.Atoi(departmentId)
	if err != nil {
		logx.Error("failed to convert department id", logx.Field("detail", err))
		return 0, errorx.NewInvalidArgumentError("failed to get department ID")
	}
	return uint64(id), nil
}
