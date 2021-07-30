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
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/drifterz13/tw-reporter-cli/api"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type TaskReportRow struct {
	TasklistTitle string
	TaskTitle     string
	TaskPoints    int
	TaskAssignees string // ex. 'member1,member2'
}

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report project status",
	Long:  `Report project status`,
	Run: func(cmd *cobra.Command, args []string) {
		// reportAll()
		// reportByTasklist()
		// reporyByUser()
		reportByUserCon()
	},
}

func reportByUserCon() {
	usersChan := make(chan []api.User, 1)
	tasklistsChan := make(chan []api.Tasklist, 1)

	go api.GetWorkspaceUsers(usersChan)
	go api.GetTasklistsCon(tasklistsChan)

	users := <-usersChan

	var items []string
	for _, user := range users {
		fullname := user.FirstName + " " + user.LastName
		items = append(items, fullname)
	}

	prompt := promptui.Select{
		Label: "Select a user",
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("prompt failed: %v\n", err)
	}

	userIdMapFullname := map[string]string{}
	for _, user := range users {
		userIdMapFullname[user.ID] = user.FirstName + " " + user.LastName
	}
	var selectedUserId string
	for k, v := range userIdMapFullname {
		if v != result {
			continue
		}

		selectedUserId = k
	}

	fmt.Printf("selected user: %v, id: %v\n", userIdMapFullname[selectedUserId], selectedUserId)

	tasklists := <-tasklistsChan
	close(tasklistsChan)

	var wg sync.WaitGroup
	taskWithTasklistChan := make(chan []api.TaskWithTasklist)

	startFetchTasksTime := time.Now()
	for _, tasklist := range tasklists {
		if tasklist.IsDeleted {
			continue
		}

		list := tasklist
		wg.Add(1)

		go api.GetTasksCon(&list, taskWithTasklistChan, &wg)
	}

	go func() {
		wg.Wait()
		close(taskWithTasklistChan)
	}()

	var allTasksWithTasklist []api.TaskWithTasklist
	for taskWithTasklist := range taskWithTasklistChan {
		for _, t := range taskWithTasklist {
			allTasksWithTasklist = append(allTasksWithTasklist, t)
		}
	}
	fmt.Printf("fetch tasks duration: %v\n", time.Since(startFetchTasksTime))

	var userTasksWithTasklist []api.TaskWithTasklist
	for _, t := range allTasksWithTasklist {
		for _, member := range t.Task.Members {
			if member.IsAssignee && member.ID == selectedUserId {
				userTasksWithTasklist = append(userTasksWithTasklist, t)
			}
		}
	}

	var tableData [][]string
	for _, t := range userTasksWithTasklist {
		var row []string
		row = append(row, t.List.Title, t.Task.Title, strconv.Itoa(t.Task.Points), userIdMapFullname[selectedUserId])
		tableData = append(tableData, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Tasklist", "Task", "Points", "Assignees"})
	table.SetRowLine(true)
	table.AppendBulk(tableData)
	table.Render()
}

func reportAll() {
	fmt.Println("getting tasklists...")
	tasklists := api.GetTasklists()

	var tableData [][]string

	usersChan := make(chan []api.User, 1)

	fmt.Println("getting users...")
	go api.GetWorkspaceUsers(usersChan)

	taskReportRow := map[string]TaskReportRow{}
	var allTasks []api.Task

	fmt.Println("getting tasks...")
	for _, tasklist := range tasklists {
		tasks := api.GetTasks(&tasklist)
		for _, task := range tasks {
			allTasks = append(allTasks, task)
			taskReportRow[task.ID] = TaskReportRow{
				TasklistTitle: tasklist.Title,
				TaskTitle:     task.Title,
				TaskPoints:    task.Points,
			}
		}
	}

	users := <-usersChan

	for _, task := range allTasks {
		assignees := api.GetTaskAssignees(task, users)
		row := taskReportRow[task.ID]
		row.TaskAssignees = assignees
		taskReportRow[task.ID] = row
	}

	for _, v := range taskReportRow {
		var row []string
		row = append(row, v.TasklistTitle, v.TaskTitle, strconv.Itoa(v.TaskPoints), v.TaskAssignees)
		tableData = append(tableData, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Tasklist", "Task", "Points", "Assignees"})
	table.SetRowLine(true)
	table.AppendBulk(tableData)
	table.Render()
}

func reportByTasklist() {
	usersChan := make(chan []api.User, 1)
	go api.GetWorkspaceUsers(usersChan)

	fmt.Println("getting tasklists...")
	tasklists := api.GetTasklists()

	var items []string

	for _, tasklist := range tasklists {
		items = append(items, tasklist.Title)
	}

	prompt := promptui.Select{
		Label: "Select a tasklist",
		Items: items,
	}

	index, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("prommpt failed %v\n", err)
	}

	fmt.Printf("select: %v, index: %v\n", result, index)

	var tableData [][]string

	tasklist := &tasklists[index]
	tasks := api.GetTasks(tasklist)
	users := <-usersChan

	for _, task := range tasks {
		var row []string

		assignees := api.GetTaskAssignees(task, users)
		row = append(row, tasklist.Title, task.Title, strconv.Itoa(task.Points), assignees)
		tableData = append(tableData, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Tasklist", "Task", "Points", "Assignees"})
	table.SetRowLine(true)
	table.AppendBulk(tableData)
	table.Render()

}

func init() {
	rootCmd.AddCommand(reportCmd)

}
