package dao

import (
	"../utils"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
)

var log = logrus.New()

//添加文件
func InsertFile(filePath string, reader io.Reader) error {
	writer, err := utils.CreateFile(filePath)
	if err != nil {
		return err
	}
	defer writer.Close()
	written, err := io.Copy(writer, reader)
	log.WithFields(logrus.Fields{"filePath": filePath, "written": written}).Info("写入文件字节数")
	return err
}

//删除文件，随便删除文件其上的空文件夹
func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	//将`/aaa/bbb/text.txt`变为`/aaa/bbb/`
	folder, _ := path.Split(filePath)
	//将`/aaa/bbb/`变为`/aaa/bbb`
	folder = path.Clean(folder)
	for {
		log.WithFields(logrus.Fields{"folder": folder}).Debug("检查文件夹是否为空后删除")
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return nil
		}
		err = os.Remove(folder)
		if err != nil {
			return err
		}
		//将`/aaa/bbb`变为`/aaa`
		//如果上面不将`/aaa/bbb/`变为`/aaa/bbb`
		//这里`/aaa/bbb/`依然会返回`/aaa/bbb/`
		folder = path.Dir(folder)
	}
}
