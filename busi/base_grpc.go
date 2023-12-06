/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package busi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dtm-labs/client/dtmcli"
	"net"
	"time"

	"github.com/dtm-labs/client/dtmcli/dtmimp"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/zhenlanghuo/dtm-examples/dtmutil"

	"github.com/dtm-labs/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/client/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/client/workflow"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// BusiGrpc busi service grpc address
var BusiGrpc = fmt.Sprintf("localhost:%d", BusiGrpcPort)

// DtmClient grpc client for dtm
var DtmClient dtmgpb.DtmClient

// BusiCli grpc client for busi
var BusiCli BusiClient

func retry(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	logger.Debugf("in retry interceptor")
	err := invoker(ctx, method, req, reply, cc, opts...)
	if st, _ := status.FromError(err); st != nil && st.Code() == codes.Unavailable {
		logger.Errorf("invoker return err: %v", err)
		time.Sleep(1000 * time.Millisecond)
		err = invoker(ctx, method, req, reply, cc, opts...)
	}
	return err
}

// GrpcStartup for grpc
func GrpcStartup() *grpc.Server {
	conn, err := grpc.Dial(dtmutil.DefaultGrpcServer, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(dtmgimp.GrpcClientLog))
	logger.FatalIfError(err)
	DtmClient = dtmgpb.NewDtmClient(conn)
	logger.Debugf("dtm client inited")
	// in github actions, the call is failed sometime, so add a retry
	conn1, err := grpc.Dial(BusiGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithChainUnaryInterceptor(workflow.Interceptor, retry))
	logger.FatalIfError(err)
	BusiCli = NewBusiClient(conn1)

	s := grpc.NewServer(grpc.UnaryInterceptor(dtmgimp.GrpcServerLog))
	RegisterBusiServer(s, &busiServer{})
	return s
}

// RunGrpc start to serve grpc
func RunGrpc(server *grpc.Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", BusiGrpcPort))
	logger.FatalIfError(err)
	logger.Debugf("busi grpc listening at %v", lis.Addr())
	err = server.Serve(lis)
	logger.Errorf("grpc server serve return: %v", err)
	logger.FatalIfError(err)
}

// busiServer is used to implement busi.BusiServer.
type busiServer struct {
	UnimplementedBusiServer
}

func (s *busiServer) QueryPrepared(ctx context.Context, in *ReqGrpc) (*BusiReply, error) {
	res := MainSwitch.QueryPreparedResult.Fetch()
	err := string2DtmError(res)

	return &BusiReply{Message: "a sample data"}, dtmgrpc.DtmError2GrpcError(err)
}

func (s *busiServer) TransIn(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), in.TransInResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransOut(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), in.TransOutResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransInRevert(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInRevertResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransOutRevert(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutRevertResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransInConfirm(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInConfirmResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransOutConfirm(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutConfirmResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransInTcc(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), in.TransInResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransOutTcc(ctx context.Context, in *ReqGrpc) (*BusiReply, error) {
	return &BusiReply{Message: "TransOutTcc"}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), in.TransOutResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransInXa(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, dtmgrpc.XaLocalTransaction(ctx, BusiConf, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		return sagaGrpcAdjustBalance(db, TransInUID, in.Amount, in.TransInResult)
	})
}

func (s *busiServer) TransOutXa(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, dtmgrpc.XaLocalTransaction(ctx, BusiConf, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		return sagaGrpcAdjustBalance(db, TransOutUID, in.Amount, in.TransOutResult)
	})
}

func (s *busiServer) TransInTccNested(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	tcc, err := dtmgrpc.TccFromGrpc(ctx)
	logger.FatalIfError(err)
	r := &emptypb.Empty{}
	err = tcc.CallBranch(in, BusiGrpc+"/busi.Busi/TransIn", BusiGrpc+"/busi.Busi/TransInConfirm", BusiGrpc+"/busi.Busi/TransInRevert", r)
	logger.FatalIfError(err)
	return r, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), in.TransInResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransOutHeaderYes(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	meta := dtmgimp.GetMetaFromContext(ctx, "test_header")
	if meta == "" {
		return &emptypb.Empty{}, errors.New("no header found in HeaderYes")
	}
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), in.TransOutResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransOutHeaderNo(ctx context.Context, in *ReqGrpc) (*emptypb.Empty, error) {
	meta := dtmgimp.GetMetaFromContext(ctx, "test_header")
	if meta != "" {
		return &emptypb.Empty{}, errors.New("header found in HeaderNo")
	}
	return &emptypb.Empty{}, nil
}

func (s *busiServer) TryGiveGift(ctx context.Context, in *GiveGiftReq) (rsp *GiveGiftRsp, err error) {
	giftOrderId := in.GetGiftOrderId()
	barrier := MustBarrierFromGrpc(ctx)
	err = barrier.CallWithDB(pdbGet(), func(tx *sql.Tx) error {
		return createGiftOrder(tx, giftOrderId)
	})
	if err != nil {
		return
	}
	rsp = &GiveGiftRsp{GiftOrderId: giftOrderId}
	return
}

func (s *busiServer) CommitGiveGift(ctx context.Context, in *GiveGiftReq) (rsp *emptypb.Empty, err error) {
	barrier := MustBarrierFromGrpc(ctx)
	err = barrier.CallWithDB(pdbGet(), func(tx *sql.Tx) error {
		return updateGiftOrder(tx, in.GetGiftOrderId(), GOSPaySuccess)
	})
	if err != nil {
		return
	}
	rsp = &emptypb.Empty{}
	return
}

func (s *busiServer) CancelGiveGift(ctx context.Context, in *GiveGiftReq) (rsp *emptypb.Empty, err error) {
	barrier := MustBarrierFromGrpc(ctx)
	err = barrier.CallWithDB(pdbGet(), func(tx *sql.Tx) error {
		return updateGiftOrder(tx, in.GetGiftOrderId(), GOSPayFailed)
	})
	if err != nil {
		return
	}
	rsp = &emptypb.Empty{}
	return
}

func (s *busiServer) TryPay(ctx context.Context, in *PayReq) (rsp *emptypb.Empty, err error) {
	//if time.Now().Unix()%2 == 0 {
	//	err = status.New(codes.Aborted, fmt.Sprintf("reason:%s", MainSwitch.FailureReason.Fetch())).Err()
	//	return
	//}
	barrier := MustBarrierFromGrpc(ctx)
	err = barrier.CallWithDB(pdbGet(), func(tx *sql.Tx) error {
		return createWalletOrder(tx, in.GetPayOrderId(), in.GetGiftOrderId())
	})
	if err != nil {
		return
	}
	rsp = &emptypb.Empty{}
	return
}

func (s *busiServer) CommitPay(ctx context.Context, in *PayReq) (rsp *emptypb.Empty, err error) {
	rsp = &emptypb.Empty{}
	return
}

func (s *busiServer) CancelPay(ctx context.Context, in *PayReq) (rsp *emptypb.Empty, err error) {
	rsp = &emptypb.Empty{}
	return
}

type GiftOrderStatus int

const (
	GOSInit       GiftOrderStatus = 1
	GOSPaySuccess GiftOrderStatus = 2
	GOSPayFailed  GiftOrderStatus = 3
)

func createGiftOrder(db dtmcli.DB, giftOrderId string) (err error) {
	_, err = dtmimp.DBExec(BusiConf.Driver, db, "insert into dtm_busi.gift_order (`id`, `status`)  VALUES (?,?)", giftOrderId, GOSInit)
	if err != nil {
		return
	}
	return
}

func updateGiftOrder(db dtmcli.DB, giftOrderId string, status GiftOrderStatus) (err error) {
	_, err = dtmimp.DBExec(BusiConf.Driver, db, "update dtm_busi.gift_order set status = ? where id = ?", status, giftOrderId)
	if err != nil {
		return
	}
	return
}

func createWalletOrder(db dtmcli.DB, payOrderId, giftOrderId string) (err error) {
	_, err = dtmimp.DBExec(BusiConf.Driver, db, "insert into dtm_busi.wallet_order (`id`, `biz_order_id`)  VALUES (?,?)", payOrderId, giftOrderId)
	if err != nil {
		return
	}
	return
}
