package config

import "os"

func UseTemporaryToken() bool {
	return os.Getenv("USE_TEMPORARY_TOKEN") == "1"
}
