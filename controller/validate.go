package controller

import (
	common_model "github.com/cellargalaxy/go_common/model"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/gin-gonic/gin"
)

//登录检查
func validate(ctx *gin.Context) {
	util.HttpValidate(ctx, &HttpValidate{})
}

type HttpValidate struct {
}

func (this *HttpValidate) GetSecret(c *gin.Context) string {
	return config.Config.Secret
}

func (this *HttpValidate) CreateClaims(c *gin.Context) *common_model.Claims {
	return &common_model.Claims{}
}
