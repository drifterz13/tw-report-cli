package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

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

func GetTasklistsCon(ch chan []Tasklist) {
	takslists := GetTasklists()

	ch <- takslists
}

func GetTasklists() []Tasklist {
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
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var result GetTasklistsResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	var activeTasklists []Tasklist

	for _, tasklist := range result.Tasklists {
		if !tasklist.IsDeleted {
			activeTasklists = append(activeTasklists, tasklist)
		}
	}

	return activeTasklists
}
