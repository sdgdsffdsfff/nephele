package util

import "os"
import cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"

//get running environment from environment variable 'NEPHELE_ENV'
//only support env 'uat' and 'prod'
func GetRunningEnv() string {
	var env string
	if os.Getenv("NEPHELE_ENV") == "prod" {
		env = "prod"
	} else {
		env = "uat"
	}
	return env
}

//initial cat based on env
func InitCat() {
	switch GetRunningEnv() {
	case "uat":
		cat.CAT_HOST = cat.UAT
	case "prod":
		cat.CAT_HOST = cat.PROD
	}
	cat.DOMAIN = "nephele"
}

