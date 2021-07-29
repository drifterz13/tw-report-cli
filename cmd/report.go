/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report project status",
	Long:  `Report project status`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("report called")
		tasklists := getTasklists()

		var tableData [][]string

		// Use only 3 tasklists for testing purpose.
		for _, tasklist := range tasklists {
			tasks := getTasks(&tasklist)

			for _, task := range tasks {
				row := append([]string{}, tasklist.Title, task.Title, strconv.Itoa(task.Points))
				tableData = append(tableData, row)
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Tasklist title", "Task title", "Points"})
		table.SetAutoMergeCells(true)
		table.SetRowLine(true)
		table.AppendBulk(tableData)
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}

type Tasklist struct {
	ID        string   `json:"list_id"`
	SpaceId   string   `json:"space_id"`
	ProjectId string   `json:"project_id"`
	Title     string   `json:"title"`
	IsDeleted bool     `json:"is_deleted"`
	Tasks     []string `json:"tasks"`
}

type GetTasklistsResponse struct {
	Ok        bool       `json:"ok"`
	Tasklists []Tasklist `json:"tasklists"`
}

func getTasklists() []Tasklist {
	val := map[string]string{
		"access_token": viper.GetString("accessToken"),
		"space_id":     viper.GetString("workspaceId"),
		"project_id":   viper.GetString("projectId"),
	}

	j, err := json.Marshal(&val)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(viper.GetString("apiUrl")+"/tasklist.get-all", "application/json", bytes.NewBuffer(j))
	var result GetTasklistsResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	var activeTasklists []Tasklist

	for _, tasklist := range result.Tasklists {
		if !tasklist.IsDeleted {
			activeTasklists = append(activeTasklists, tasklist)
			fmt.Printf("tasklist title: %v, contains total tasks: %v\n", tasklist.Title, len(tasklist.Tasks))
		}
	}

	return activeTasklists
}

type TaskMember struct {
	ID         string `json:"_id"`
	IsAssignee bool   `json:"is_assignee"`
}

type Task struct {
	TaskId    string       `json:"task_id"`
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

func getTask(taskId string, ch chan GetTaskResponse, wg *sync.WaitGroup) {
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

	var result GetTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	ch <- result
}

func getTasks(tasklist *Tasklist) []Task {
	var wg sync.WaitGroup
	taskChan := make(chan GetTaskResponse, len(tasklist.Tasks))

	for _, taskId := range tasklist.Tasks {
		wg.Add(1)

		go getTask(taskId, taskChan, &wg)
	}

	wg.Wait()
	close(taskChan)

	var tasks []Task
	for task := range taskChan {
		tasks = append(tasks, task.Task)
	}

	for _, task := range tasks {
		fmt.Printf("Tasklist: %v contains task: %v with points %v\n", tasklist.Title, task.Title, task.Points)
	}

	return tasks
}
