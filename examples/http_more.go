package examples

import (
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/lithammer/shortuuid/v3"
	"github.com/zhenlanghuo/dtm-examples/busi"
	"github.com/zhenlanghuo/dtm-examples/dtmutil"
)

func init() {
	AddCommand("http_saga_multiSource", func() string {
		req := &busi.ReqHTTP{Amount: 30}
		saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, shortuuid.New()).
			Add(busi.Busi+"/SagaMultiSource", busi.Busi+"/SagaMultiSourceRevert", req)
		logger.Debugf("saga busi trans submit")
		err := saga.Submit()
		logger.Debugf("result gid is: %s", saga.Gid)
		logger.FatalIfError(err)
		return saga.Gid
	})
}
