package config

import "os"

func EnableTaskEnvelope() bool {
	return os.Getenv("ENABLE_TASK_ENVELOPE") == "1"
}
