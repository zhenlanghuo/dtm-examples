/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/client/dtmcli/logger"
	dtmgrpc "github.com/dtm-labs/client/dtmgrpc"
	"github.com/lithammer/shortuuid/v3"
	"github.com/zhenlanghuo/dtm-examples/busi"
	"github.com/zhenlanghuo/dtm-examples/dtmutil"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	AddCommand("grpc_saga_customHeaders", func() string {
		req := &busi.ReqGrpc{Amount: 30}
		gid := shortuuid.New()
		saga := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gid).
			Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", req) // /TransOutHeaderYes will check header exists

		saga.BranchHeaders = map[string]string{
			"test_header": "test",
		}
		saga.WaitResult = true
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
	AddCommand("grpc_tcc_customHeaders", func() string {
		gid := shortuuid.New()
		err := dtmgrpc.TccGlobalTransaction2(dtmutil.DefaultGrpcServer, gid, func(tg *dtmgrpc.TccGrpc) {
			tg.BranchHeaders = map[string]string{
				"test_header": "test",
			}
			tg.WaitResult = true
		}, func(tcc *dtmgrpc.TccGrpc) error {
			data := &busi.ReqGrpc{Amount: 30}
			r := &emptypb.Empty{}
			return tcc.CallBranch(data, busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", "", r)
		})
		logger.FatalIfError(err)
		return gid
	})
	AddCommand("grpc_saga_passthroughHeaders", func() string {
		//dtmcli.SetPassthroughHeaders([]string{"test_header"})        // set passthrough headers. dtm will
		dtmgrpc.AddUnaryInterceptor(busi.SetGrpcHeaderForHeadersYes) // will set header in this middleware

		req := &busi.ReqGrpc{Amount: 30}
		gid := shortuuid.New() + "HeadersYes" // gid with this post fix will be handled in interceptor
		saga := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gid).
			Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", req) // /TransOutHeaderYes will check header exists

		saga.WaitResult = true
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
}
