/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/dtm-examples/busi"
	"github.com/dtm-labs/dtm-examples/dtmutil"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
)

func init() {
	AddCommand("http_saga_redis", func() string {
		busi.SetRedisBothAccount(10000, 10000)
		req := &busi.TransReq{Amount: 30}
		saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, dtmcli.MustGenGid(dtmutil.DefaultHTTPServer)).
			Add(busi.Busi+"/SagaRedisTransOut", busi.Busi+"/SagaRedisTransOutCompensate", req).
			Add(busi.Busi+"/SagaRedisTransIn", busi.Busi+"/SagaRedisTransInCompensate", req)
		logger.Debugf("busi trans submit")
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
	AddCommand("http_saga_redis_rollback", func() string {
		busi.SetRedisBothAccount(10, 10)
		req := &busi.TransReq{Amount: 30}
		saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, dtmcli.MustGenGid(dtmutil.DefaultHTTPServer)).
			Add(busi.Busi+"/SagaRedisTransIn", busi.Busi+"/SagaRedisTransInCom", req).
			Add(busi.Busi+"/SagaRedisTransOut", busi.Busi+"/SagaRedisTransOutCom", req)
		logger.Debugf("busi trans submit")
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
}