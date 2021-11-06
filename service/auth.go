package service

import "github.com/cellargalaxy/go_file_bed/config"

func CheckToken(token string) bool {
	return config.Token == token
}
