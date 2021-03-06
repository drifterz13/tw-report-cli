package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/spf13/viper"
)

type TaskMember struct {
	ID         string `json:"user_id"`
	IsAssignee bool   `json:"is_assignee"`
}

type Task struct {
	ID        string       `json:"task_id"`
	OwnerId   string       `json:"owner_id"`
	SpaceId   string       `json:"space_id"`
	Title     string       `json:"title"`
	Members   []TaskMember `json:"members"`
	Tags      []string     `json:"tags"`
	Points    int          `json:"points"`
	Status    int          `json:"status"`
	IsDeleted bool         `json:"is_deleted"`
}

type GetTaskResponse struct {
	Ok   bool `json:"ok"`
	Task Task `json:"task"`
}

func GetTask(taskId string, ch chan GetTaskResponse, wg *sync.WaitGroup) {
	defer wg.Done()
	val := map[string]string{
		"access_token": viper.GetString("accessToken"),
		"space_id":     viper.GetString("workspaceId"),
		"task_id":      taskId,
	}
	j, err := json.Marshal(&val)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(viper.GetString("apiUrl")+"/task.get", "application/json", bytes.NewBuffer(j))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var result GetTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	ch <- result
}

type TaskWithTasklist struct {
	Task
	List Tasklist
}

func GetTasksWithTasklist(tasklist *Tasklist, ch chan []TaskWithTasklist, wg *sync.WaitGroup) {
	defer wg.Done()
	tasks := GetTasks(tasklist)

	fmt.Printf("received `%v` tasks from tasklist: %v\n", len(tasks), tasklist.Title)

	var t []TaskWithTasklist
	for _, task := range tasks {
		r := TaskWithTasklist{
			Task: task,
			List: *tasklist,
		}
		t = append(t, r)
	}

	ch <- t
}

func GetTasks(tasklist *Tasklist) []Task {
	var wg sync.WaitGroup
	taskChan := make(chan GetTaskResponse, len(tasklist.Tasks))

	for _, taskId := range tasklist.Tasks {
		wg.Add(1)

		go GetTask(taskId, taskChan, &wg)
	}

	wg.Wait()
	close(taskChan)

	var tasks []Task
	for task := range taskChan {
		tasks = append(tasks, task.Task)
	}

	return tasks
}
