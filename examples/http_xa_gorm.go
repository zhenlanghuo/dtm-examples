/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/go-resty/resty/v2"
	"github.com/lithammer/shortuuid/v3"
	"github.com/zhenlanghuo/dtm-examples/busi"
	"github.com/zhenlanghuo/dtm-examples/dtmutil"
)

func init() {
	AddCommand("http_xa_gorm", func() string {
		gid := shortuuid.New()
		err := dtmcli.XaGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
			resp, err := xa.CallBranch(&busi.ReqHTTP{Amount: 30}, busi.Busi+"/TransOutXaGorm")
			if err != nil {
				return resp, err
			}
			return xa.CallBranch(&busi.ReqHTTP{Amount: 30}, busi.Busi+"/TransInXa")
		})
		logger.FatalIfError(err)
		return gid
	})

}
