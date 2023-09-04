package examples

import (
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/lithammer/shortuuid/v3"
	"github.com/zhenlanghuo/dtm-examples/busi"
	"github.com/zhenlanghuo/dtm-examples/dtmutil"
)

func init() {
	AddCommand("http_saga_mongo", func() string {
		busi.SetMongoBothAccount(10000, 10000)
		req := &busi.ReqHTTP{Amount: 30}
		saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, shortuuid.New()).
			Add(busi.Busi+"/SagaMongoTransOut", busi.Busi+"/SagaMongoTransOutCom", req).
			Add(busi.Busi+"/SagaMongoTransIn", busi.Busi+"/SagaMongoTransInCom", req)
		logger.Debugf("busi trans submit")
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
	AddCommand("http_saga_mongo_rollback", func() string {
		busi.SetMongoBothAccount(10000, 10000)
		saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, shortuuid.New()).
			Add(busi.Busi+"/SagaMongoTransIn", busi.Busi+"/SagaMongoTransInCom", &busi.ReqHTTP{Amount: 30}).
			Add(busi.Busi+"/SagaMongoTransOut", busi.Busi+"/SagaMongoTransOutCom", &busi.ReqHTTP{Amount: 30000})
		logger.Debugf("busi trans submit")
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
}
