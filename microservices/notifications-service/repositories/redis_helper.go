package repositories

import (
	"fmt"
)

const (
	cacheProjects = "notifications:%s"
)

func constructKeyProjects(username string) string {
	return fmt.Sprintf(cacheProjects, username)
}
