package repository

import (
	"fmt"
)

const (
	cacheTask  = "task:%s"
	cacheTasks = "tasks:%s"
)

func constructKeyOneProject(id string) string {
	return fmt.Sprintf(cacheTask, id)
}

func constructKeyProjects(projectId string) string {
	return fmt.Sprintf(cacheTasks, projectId)
}
