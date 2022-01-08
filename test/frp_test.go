package test

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/service"
	"testing"
)

func TestFrpPull(test *testing.T) {
	ctx := util.CreateLogCtx()
	err := service.FrpPull(ctx, "http://192.168.123.5:7090", "admin", "", "", "")
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}
