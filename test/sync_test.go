package test

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/service"
	"testing"
)

func TestPushFile(test *testing.T) {
	ctx := util.GenCtx()
	err := service.PushSyncFile(ctx, "http://127.0.0.1:8880/", "secret", "")
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}

func TestPullFile(test *testing.T) {
	ctx := util.GenCtx()
	err := service.PullSyncFile(ctx, "http://127.0.0.1:8880/", "secret", "")
	if err != nil {
		test.Error(err)
		test.FailNow()
	}
}
