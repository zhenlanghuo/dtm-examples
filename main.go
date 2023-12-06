package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dtm-labs/client/dtmcli/dtmimp"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/dtm-labs/client/workflow"
	"github.com/zhenlanghuo/dtm-examples/busi"
	"github.com/zhenlanghuo/dtm-examples/dtmutil"
	"github.com/zhenlanghuo/dtm-examples/examples"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func hintExit(msg string) {
	if msg != "" {
		fmt.Print(msg, "\n")
	}
	fmt.Printf("Usage: %s <command>\n\nCommand can be one of the following:\n\n", filepath.Base(os.Args[0]))
	fmt.Printf("%4s%-32srun a quick start example\n", "", "qs")
	for _, cmd := range examples.Commands {
		fmt.Printf("%4s%-32srun an example includes %s\n", "", cmd.Arg, strings.ReplaceAll(cmd.Arg, "_", " "))
	}
	os.Exit(0)
}
func main() {
	if len(os.Args) == 1 {
		hintExit("")
	}
	logger.InitLog("debug")
	//busi.StoreHost = "en.dtm.pub"
	busi.StoreHost = "127.0.0.1"
	busi.BusiConf = dtmimp.DBConf{
		Driver: "mysql",
		Host:   busi.StoreHost,
		Port:   3306,
		//User:     "dtm",
		//Password: "passwd123dtm",
		User:     "root",
		Password: "123456",
	}
	busi.ResetXaData()

	cmd := os.Args[1]
	app, gsvr := busi.Startup(cmd)
	examples.AddRoutes(app)
	time.Sleep(200 * time.Millisecond)
	if cmd == "http_workflow_tcc_barrier_TccBTransInTryFailed_TccBTransOutCancelTimeout" {
		cmd = "http_workflow_tcc_barrier"
	}
	if cmd == "qs" {
		go busi.RunHTTP(app)
		time.Sleep(200 * time.Millisecond)
		busi.QsMain()
	} else if examples.IsExists(cmd) {
		if strings.Contains(cmd, "grpc") { // init workflow base on command
			nossl := grpc.WithTransportCredentials(insecure.NewCredentials())
			workflow.InitGrpc(dtmutil.DefaultGrpcServer, busi.BusiGrpc, gsvr)
			conn1, err := grpc.Dial(busi.BusiGrpc, grpc.WithUnaryInterceptor(workflow.Interceptor), nossl)
			logger.FatalIfError(err)
			busi.BusiCli = busi.NewBusiClient(conn1)
		} else {
			workflow.InitHTTP(dtmutil.DefaultHTTPServer, busi.Busi+"/workflow/resume")
		}
		go busi.RunGrpc(gsvr)
		go busi.RunHTTP(app)
		time.Sleep(200 * time.Millisecond)
		examples.Call(cmd)
	} else {
		hintExit("unknown command: " + cmd)
	}
	select {}
}
