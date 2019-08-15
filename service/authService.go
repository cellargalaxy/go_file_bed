package service

import (
	"../config"
)

func CheckToken(token string) bool {
	return config.GetConfig().Token == token
}
