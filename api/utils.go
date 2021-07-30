package api

import (
	"strings"
)

func GetTaskAssignees(task Task, users []User) string {
	var taskAssinees []string
	userIdMapFirstName := map[string]string{}

	for _, user := range users {
		userIdMapFirstName[user.ID] = user.FirstName
	}

	for _, member := range task.Members {
		if member.IsAssignee {
			assignee := userIdMapFirstName[member.ID]
			taskAssinees = append(taskAssinees, assignee)
		}
	}

	assignees := strings.Join(taskAssinees, ",")
	return assignees
}
