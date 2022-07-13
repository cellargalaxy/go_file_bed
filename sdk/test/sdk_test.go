package test

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/cellargalaxy/go_file_bed/sdk"
	"testing"
)

const (
	address = "http://127.0.0.1" + model.ListenAddress
	secret  = "secret"
)

func TestAddFile(test *testing.T) {
	ctx := util.GenCtx()
	client, err := sdk.NewDefaultFileBedClient(ctx, address, secret)
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
	filePath := ""
	reader, err := util.GetReadFile(ctx, "")
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
	response, err := client.AddFile(ctx, filePath, reader, true)
	test.Logf("response: %+v\r\n", util.ToJsonIndentString(response))
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}

func TestGetFileCompleteInfo(test *testing.T) {
	ctx := util.GenCtx()
	client, err := sdk.NewDefaultFileBedClient(ctx, address, secret)
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
	var request model.FileCompleteInfoGetRequest
	request.Path = "aaa/20211205/(pid-42733520)Forever.png.JPEG"
	response, err := client.GetFileCompleteInfo(ctx, request)
	test.Logf("response: %+v\r\n", util.ToJsonIndentString(response))
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}

func TestListFileSimpleInfo(test *testing.T) {
	ctx := util.GenCtx()
	client, err := sdk.NewDefaultFileBedClient(ctx, address, secret)
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
	var request model.FileSimpleInfoListRequest
	request.Path = "aaa/20211205"
	response, err := client.ListFileSimpleInfo(ctx, request)
	test.Logf("response: %+v\r\n", util.ToJsonIndentString(response))
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}
