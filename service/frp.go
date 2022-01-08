package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/dao"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strings"
)

func FrpPull(ctx context.Context, address, username, password, localPath, remotePath string) error {
	client, err := NewFrpClient(ctx, address, username, password)
	if err != nil {
		return err
	}
	return client.Pull(ctx, localPath, remotePath)
}

type FrpHandler struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"-"`
}

func (this FrpHandler) String() string {
	return util.ToJsonString(this)
}

func (this FrpHandler) GetAddress(ctx context.Context) string {
	if strings.HasSuffix(this.Address, "/") {
		return this.Address[:len(this.Address)-1]
	}
	return this.Address
}
func (this FrpHandler) GetUsername(ctx context.Context) string {
	return this.Username
}
func (this FrpHandler) GetPassword(ctx context.Context) string {
	return this.Password
}

func NewFrpClient(ctx context.Context, address, username, password string) (model.FrpInter, error) {
	var handler FrpHandler
	handler.Address = address
	handler.Username = username
	handler.Password = password
	var client FrpClient
	client.handler = handler
	return client, nil
}

type FrpClient struct {
	handler model.FrpHandlerInter
}

func (this FrpClient) Pull(ctx context.Context, localPath, remotePath string) error {
	tmpInfo, err := this.requestFrpFile(ctx, remotePath)
	if tmpInfo == nil {
		return err
	}
	if err != nil {
		dao.DeleteFile(ctx, tmpInfo.Path)
		return err
	}
	completeInfo, err := GetFileCompleteInfo(ctx, tmpInfo.Path)
	if err != nil {
		dao.DeleteFile(ctx, tmpInfo.Path)
		return err
	}

	if completeInfo.Size >= 1024*1024 { //1M
		err = dao.MoveFile(ctx, tmpInfo.Path, localPath)
		if err != nil {
			dao.DeleteFile(ctx, localPath)
			dao.DeleteFile(ctx, tmpInfo.Path)
		}
		return err
	}

	data, err := dao.GetFileData(ctx, tmpInfo.Path)
	if err != nil {
		dao.DeleteFile(ctx, tmpInfo.Path)
		return err
	}
	names, err := this.parseFrpFile(ctx, data)
	if err != nil {
		err = dao.MoveFile(ctx, tmpInfo.Path, localPath)
		if err != nil {
			dao.DeleteFile(ctx, localPath)
			dao.DeleteFile(ctx, tmpInfo.Path)
		}
		return err
	}
	dao.DeleteFile(ctx, tmpInfo.Path)
	for i := range names {
		localSubPath := path.Join(localPath, names[i])
		remoteSubPath := path.Join(remotePath, names[i])
		this.Pull(ctx, localSubPath, remoteSubPath)
	}
	return nil
}

func (this FrpClient) requestFrpFile(ctx context.Context, remotePath string) (*model.FileSimpleInfo, error) {
	if !strings.HasPrefix(remotePath, "/") {
		remotePath = "/" + remotePath
	}
	url := this.handler.GetAddress(ctx) + remotePath
	logrus.WithContext(ctx).WithFields(logrus.Fields{"url": url}).Info("frp文件")

	response, err := httpClient.R().SetContext(ctx).
		SetBasicAuth(this.handler.GetUsername(ctx), this.handler.GetPassword(ctx)).
		SetHeader("User-Agent", userAgent).
		SetDoNotParseResponse(true).
		Get(url)

	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("frp文件，文件下载异常")
		return nil, fmt.Errorf("frp文件，文件下载异常")
	}
	if response == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("frp文件，文件下载响应为空")
		return nil, fmt.Errorf("frp文件，文件下载响应为空")
	}
	statusCode := response.StatusCode()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode}).Info("frp文件，文件下载响应")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"StatusCode": statusCode}).Error("frp文件，文件下载响应码失败")
		return nil, fmt.Errorf("frp文件，文件下载响应码失败: %+v", statusCode)
	}

	reader := response.RawBody()
	defer reader.Close()
	return AddTmpFile(ctx, reader)
}

type pre struct {
	XMLName xml.Name `xml:"pre"`
	Names   []string `xml:"a"`
}

func (this pre) String() string {
	return util.ToJsonString(this)
}

func (this FrpClient) parseFrpFile(ctx context.Context, data []byte) ([]string, error) {
	var p pre
	err := xml.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"p": p}).Info("frp文件")
	return p.Names, nil
}
