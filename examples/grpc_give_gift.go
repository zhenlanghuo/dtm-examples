package examples

import (
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/lithammer/shortuuid/v3"
	"github.com/zhenlanghuo/dtm-examples/busi"
	"github.com/zhenlanghuo/dtm-examples/dtmutil"
	"google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	AddCommand("grpc_tcc_give_gift", func() string {
		logger.Debugf("tcc simple transaction begin")
		gid := shortuuid.New()
		err := dtmgrpc.TccGlobalTransaction(dtmutil.DefaultGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
			giveGiftReq := &busi.GiveGiftReq{GiftOrderId: gid}
			giveGiftRsp := &busi.GiveGiftRsp{}
			err := tcc.CallBranch(giveGiftReq, busi.BusiGrpc+"/busi.Busi/TryGiveGift", busi.BusiGrpc+"/busi.Busi/CommitGiveGift",
				busi.BusiGrpc+"/busi.Busi/CancelGiveGift", giveGiftRsp)
			if err != nil {
				return err
			}
			logger.Debugf("grpc_tcc_give_gift, giveGiftRsp: %v", giveGiftRsp)
			//logger.FatalIfError(errors.New("测试中途失败"))
			payReq := &busi.PayReq{
				PayOrderId:  gid,
				GiftOrderId: gid,
			}
			payRsp := &emptypb.Empty{}
			err = tcc.CallBranch(payReq, busi.BusiGrpc+"/busi.Busi/TryPay", busi.BusiGrpc+"/busi.Busi/CommitPay", busi.BusiGrpc+"/busi.Busi/CancelPay", payRsp)
			return err
		})
		logger.FatalIfError(err)
		return gid
	})
}
