package util

import "os"

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
