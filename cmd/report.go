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
	"os"
	"strconv"

	"github.com/drifterz13/tw-reporter-cli/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report project status",
	Long:  `Report project status`,
	Run: func(cmd *cobra.Command, args []string) {
		tasklists := api.GetTasklists()

		var tableData [][]string

		usersChan := make(chan []api.User, 1)
		go api.GetWorkspaceUsers(usersChan)

		users := <-usersChan
		userIdMapFirstName := map[string]string{}

		for _, user := range users {
			userIdMapFirstName[user.ID] = user.FirstName
		}

		for _, tasklist := range tasklists[:3] {
			tasks := api.GetTasks(&tasklist)
			for _, task := range tasks {
				for _, member := range task.Members {
					var taskAssignee string
					if member.IsAssignee {
						taskAssignee = userIdMapFirstName[member.ID]
					}
					row := append([]string{}, tasklist.Title, task.Title, taskAssignee, strconv.Itoa(task.Points))
					tableData = append(tableData, row)
				}

			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Tasklist", "Task", "Assignee", "Points"})
		table.SetAutoMergeCells(true)
		table.SetRowLine(true)
		table.AppendBulk(tableData)
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
