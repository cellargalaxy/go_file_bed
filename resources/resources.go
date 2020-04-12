package resources

import (
	rice "github.com/GeertJohan/go.rice"
	_ "github.com/cellargalaxy/go-file-bed/docs"
)

var StaticBox *rice.Box

func init() {
	StaticBox = rice.MustFindBox("static")
}
