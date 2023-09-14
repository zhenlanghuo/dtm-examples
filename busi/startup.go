package busi

import (
	"database/sql"
	"fmt"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	grpc "google.golang.org/grpc"
)

// Startup startup the busi's grpc and http service
func Startup(cmd string) (*gin.Engine, *grpc.Server) {
	if cmd == "http_workflow_tcc_barrier_TccBTransInTryFailed_TccBTransOutCancelTimeout" {
		TccBarrierTransOutCancel = func(c *gin.Context) interface{} {
			req := reqFrom(c)
			bb := MustBarrierFromGin(c)
			if req.Store == Redis {
				return bb.RedisCheckAdjustAmount(RedisGet(), GetRedisAccountKey(TransOutUID), -req.Amount, 7*86400)
			}
			if req.Store == Mongo {
				return bb.MongoCall(MongoGet(), func(sc mongo.SessionContext) error {
					return SagaMongoAdjustBalance(sc, sc.Client(), TransOutUID, reqFrom(c).Amount, "")
				})
			}

			err := bb.CallWithDB(pdbGet(), func(tx *sql.Tx) error {
				return tccAdjustTrading(tx, TransOutUID, reqFrom(c).Amount)
			})
			if err != nil {
				return err
			}
			return fmt.Errorf("返回普通错误, 模拟超时错误的效果")
		}

		TccBarrierTansInTry = func(c *gin.Context) interface{} {
			req := reqFrom(c)
			if req.TransInResult != "" {
				return string2DtmError(req.TransInResult)
			}
			return MustBarrierFromGin(c).CallWithDB(pdbGet(), func(tx *sql.Tx) error {
				// 假设 TccBTransInTry 是扣金币，此时金币不足，返回失败
				fmt.Println("flag!!")
				if req.Amount >= 30 {
					return dtmcli.ErrFailure
				}
				return tccAdjustTrading(tx, TransInUID, req.Amount)
			})
		}
	}

	svr := GrpcStartup()
	app := BaseAppStartup()
	return app, svr
}
