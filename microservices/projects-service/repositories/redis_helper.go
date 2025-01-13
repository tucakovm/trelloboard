package repositories

import (
	"fmt"
)

const (
	cacheProject  = "project:%s"
	cacheProjects = "projects:%s"
)

func constructKeyOneProject(id string) string {
	return fmt.Sprintf(cacheProject, id)
}

func constructKeyProjects(username string) string {
	return fmt.Sprintf(cacheProjects, username)
}
