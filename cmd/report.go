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

type TaskReport struct {
	header []string
	data   [][]string
}

func (tr *TaskReport) addRow(row []string) {
	tr.data = append(tr.data, row)
}

func (tr *TaskReport) printTable() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tr.header)
	table.SetRowLine(true)
	table.AppendBulk(tr.data)
	table.Render()
}

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report project status",
	Long:  `Report project status`,
	Run: func(cmd *cobra.Command, args []string) {
		allOption := "All"
		tasklistOption := "By Tasklist"
		userOption := "By User"

		usersChan := make(chan []api.User, 1)
		tasklistChan := make(chan []api.Tasklist, 1)

		go api.GetWorkspaceUsers(usersChan)
		go api.GetTasklists(tasklistChan)

		prompt := promptui.Select{
			Label: "Select a report option",
			Items: []string{allOption, tasklistOption, userOption},
		}

		_, result, err := prompt.Run()
		if err != nil {
			log.Fatalf("prompt failed: %v\n", err)
		}

		switch result {
		case allOption:
			reportAll(usersChan, tasklistChan)
		case tasklistOption:
			reportByTasklist(usersChan, tasklistChan)
		case userOption:
			reportByUser(usersChan, tasklistChan)
		default:
			fmt.Printf("Unknow option: %v\n", result)
		}

	},
}

func reportByUser(usersCh <-chan []api.User, tasklistCh <-chan []api.Tasklist) {
	users := <-usersCh

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

	tasklists := <-tasklistCh

	var wg sync.WaitGroup
	taskWithTasklistChan := make(chan []api.TaskWithTasklist)

	startFetchTasksTime := time.Now()

	for _, tasklist := range tasklists {
		if tasklist.IsDeleted {
			continue
		}

		wg.Add(1)
		go func(tasklist api.Tasklist) {
			api.GetTasksCon(&tasklist, taskWithTasklistChan, &wg)
		}(tasklist)
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

	taskReport := &TaskReport{
		header: []string{"Tasklist", "Task", "Points", "Assignees"},
	}
	for _, t := range userTasksWithTasklist {
		var row []string
		row = append(row, t.List.Title, t.Task.Title, strconv.Itoa(t.Task.Points), userIdMapFullname[selectedUserId])
		taskReport.addRow(row)
	}

	taskReport.printTable()
}

func reportAll(usersCh <-chan []api.User, tasklistCh <-chan []api.Tasklist) {
	tasklists := <-tasklistCh

	var wg sync.WaitGroup
	taskWithTasklistChan := make(chan []api.TaskWithTasklist)

	for _, tasklist := range tasklists {
		if tasklist.IsDeleted {
			continue
		}

		wg.Add(1)
		go func(tasklist api.Tasklist) {
			api.GetTasksCon(&tasklist, taskWithTasklistChan, &wg)
		}(tasklist)
	}

	go func() {
		wg.Wait()
		close(taskWithTasklistChan)
	}()

	users := <-usersCh

	taskReport := &TaskReport{
		header: []string{"Tasklist", "Task", "Points", "Assignees"},
	}

	var allTaskWithTasklist []api.TaskWithTasklist
	for t := range taskWithTasklistChan {
		allTaskWithTasklist = append(allTaskWithTasklist, t...)
	}

	for _, t := range allTaskWithTasklist {
		row := []string{t.List.Title, t.Task.Title, strconv.Itoa(t.Task.Points), api.GetTaskAssignees(t.Task, users)}
		taskReport.addRow(row)
	}

	taskReport.printTable()
}

func reportByTasklist(usersCh <-chan []api.User, tasklistCh <-chan []api.Tasklist) {
	tasklists := <-tasklistCh

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

	taskReport := &TaskReport{
		header: []string{"Tasklist", "Task", "Points", "Assignees"},
	}

	tasklist := &tasklists[index]
	tasks := api.GetTasks(tasklist)
	users := <-usersCh

	for _, task := range tasks {
		var row []string
		assignees := api.GetTaskAssignees(task, users)
		row = append(row, tasklist.Title, task.Title, strconv.Itoa(task.Points), assignees)
		taskReport.addRow(row)
	}

	taskReport.printTable()

}

func init() {
	rootCmd.AddCommand(reportCmd)

}
