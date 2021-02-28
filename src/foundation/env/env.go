package env

import (
	"fmt"
	"os"
)

func GetEnv(key string) string {
	val, found := os.LookupEnv(key)
	if !found {
		panic((fmt.Errorf("environment variable %v not found", key)))
	}

	return val
}

const (
	CRAWLER_TASKS_QUEUE      = "CRAWLER_TASK_QUEUE"
	CRAWLER_MACHINE_ARN      = "CRAWLER_MACHINE_ARN"
	CRAWLER_EXECUTIONS_QUEUE = "CRAWLER_EXECUTIONS_QUEUE"
)
