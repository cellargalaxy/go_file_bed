package main

import (
	"./controller"
	"./staticBuild"
	"./utils"
	"fmt"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	err := utils.OutputStaticFile(staticBuild.Base64Map)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Panic("静态文件输出失败")
	}
	log.Info("静态文件输出成功")
}

func main() {
	controller.Controller()
}

func buildStaticFile() {
	fmt.Println(utils.BuildStaticFile("goFileBed.html"))
}
