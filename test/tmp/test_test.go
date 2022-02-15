package tmp

import (
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"path"
	"strconv"
	"strings"
	"testing"
)

func TestTest(test *testing.T) {
	ctx := util.CreateLogCtx()

	fmt.Println(util.ClearPath(ctx, path.Join("/aaa/", model.TrashPath)))

	file := "/aaa/bbb/ccc.ddd"
	file = util.ClearPath(ctx, file)
	ext := path.Ext(file)
	p := strings.TrimRight(file, ext)
	fmt.Println(file)
	fmt.Println(p)
	fmt.Println(ext)
	ligId := util.GetLogId(ctx)
	fmt.Printf("%+v.%+v%+v\n", p, strconv.Itoa(int(ligId)), ext)
}
