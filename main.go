package main

import (
	"./controller"
	"./service"
	"./static"
	"./utils"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()

func init() {
	err := utils.OutputStaticFile(static.Base64Map)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Panic("静态文件输出失败")
	}
	log.Info("静态文件输出成功")
}

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
		log.WithFields(logrus.Fields{"cmd": cmd}).Info("命令入参")
		if cmd == "syn" {
			log.Info("同步文件模式")
			service.SynFile()
		} else {
			log.WithFields(logrus.Fields{"cmd": cmd}).Error("非法命令入参")
		}
		return
	}
	log.Info("web模式")
	controller.Controller()
}

func buildStaticFile() {
	fmt.Println(utils.BuildStaticFile("templates"))
}
