package test

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/service"
	"testing"
)

func TestPushFile(test *testing.T) {
	ctx := util.CreateLogCtx()
	err := service.PushSyncFile(ctx, "", "", "")
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}

func TestPullFile(test *testing.T) {
	ctx := util.CreateLogCtx()
	err := service.PullSyncFile(ctx, "", "", "")
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}
