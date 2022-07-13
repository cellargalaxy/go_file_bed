package main

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/controller"
	"github.com/cellargalaxy/go_file_bed/corn"
	"github.com/cellargalaxy/go_file_bed/model"
)

func init() {
	ctx := util.GenCtx()
	util.Init(model.DefaultServerName)
	corn.Init(ctx)
}

/**
export server_name=go_file_bed
export server_center_address=http://127.0.0.1:7557
export server_center_secret=secret_secret

server_name=go_file_bed;server_center_address=http://127.0.0.1:7557;server_center_secret=secret_secret
*/
func main() {
	err := controller.Controller()
	if err != nil {
		panic(err)
	}
}
