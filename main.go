package main

import (
	"./controller"
	"./service"
	"./staticBuild"
	"./utils"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
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
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
		log.WithFields(logrus.Fields{"cmd": cmd}).Info("命令入参")
		if cmd == "syn" {
			log.Info("同步文件模式")
			if err := service.SynFile(); err != nil {
				log.WithFields(logrus.Fields{"err": err}).Error("同步文件失败")
			}
		} else {
			log.WithFields(logrus.Fields{"cmd": cmd}).Error("非法命令入参")
		}
		return
	}
	log.Info("web模式")
	controller.Controller()
}

func buildStaticFile() {
	fmt.Println(utils.BuildStaticFile("goFileBed.html"))
}
