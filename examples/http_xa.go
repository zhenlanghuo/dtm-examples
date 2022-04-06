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
	"github.com/go-resty/resty/v2"
)

func init() {
	AddCommand("http_xa", func() string {
		gid := dtmcli.MustGenGid(dtmutil.DefaultHTTPServer)
		err := dtmcli.XaGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
			resp, err := xa.CallBranch(&busi.TransReq{Amount: 30}, busi.Busi+"/TransOutXa")
			if err != nil {
				return resp, err
			}
			return xa.CallBranch(&busi.TransReq{Amount: 30}, busi.Busi+"/TransInXa")
		})
		logger.FatalIfError(err)
		return gid
	})
}
