package service

import "github.com/cellargalaxy/go-file-bed/config"

func CheckToken(token string) bool {
	return config.Token == token
}
