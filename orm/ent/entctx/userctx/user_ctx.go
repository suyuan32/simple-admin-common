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

package userctx

import (
	"context"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

// GetUserIDFromCtx returns user id from context.
func GetUserIDFromCtx(ctx context.Context) (string, error) {
	if userId, ok := ctx.Value("userId").(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get user id from context", logx.Field("detail", ctx))
			return "", errorx.NewInvalidArgumentError("failed to get user id from context")
		} else {
			if data := md.Get(enum.UserIdRpcCtxKey); len(data) > 0 {
				userId = data[0]
				return userId, nil
			} else {
				return "", errorx.NewInvalidArgumentError("failed to get user id from context")
			}
		}
	} else {
		return userId, nil
	}
}
