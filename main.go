package main

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/cellargalaxy/go_file_bed/controller"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(config.Config.LogLevel)
	util.InitLog(util.GetServerNameWithPanic())
}

/**
export server_name=go_file_bed
export server_center_address=http://172.17.0.4:7557
export server_center_secret=secret_secret

server_name=go_file_bed;server_center_address=http://172.17.0.4:7557;server_center_secret=secret_secret
*/
func main() {
	err := controller.Controller()
	if err != nil {
		panic(err)
	}
}
