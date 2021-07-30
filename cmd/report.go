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
	"os"
	"strconv"

	"github.com/drifterz13/tw-reporter-cli/api"
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
		fmt.Println("getting tasklists...")
		tasklists := api.GetTasklists()

		var tableData [][]string

		usersChan := make(chan []api.User, 1)

		fmt.Println("getting users...")
		go api.GetWorkspaceUsers(usersChan)

		users := <-usersChan

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

		for _, task := range allTasks {
			assignees := api.GetTaskAssignees(task, users)
			r := taskReportRow[task.ID]
			taskReportRow[task.ID] = TaskReportRow{
				TasklistTitle: r.TasklistTitle,
				TaskTitle:     r.TaskTitle,
				TaskPoints:    r.TaskPoints,
				TaskAssignees: assignees,
			}
		}

		for _, v := range taskReportRow {
			row := append([]string{}, v.TasklistTitle, v.TaskTitle, strconv.Itoa(v.TaskPoints), v.TaskAssignees)
			tableData = append(tableData, row)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Tasklist", "Task", "Points", "Assignees"})
		table.SetRowLine(true)
		table.AppendBulk(tableData)
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
