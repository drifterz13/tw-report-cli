package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

type User struct {
	ID        string `json:"_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type GetWorkspaceUsersResponse struct {
	Ok    bool   `json:"ok"`
	Users []User `json:"users"`
}

func GetWorkspaceUsers(ch chan []User) {
	apiUrl := viper.GetString("apiUrl")
	accessToken := viper.GetString("accessToken")
	workspaceId := viper.GetString("workspaceId")

	val := map[string]string{
		"access_token": accessToken,
		"space_id":     workspaceId,
	}

	j, err := json.Marshal(&val)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(apiUrl+"/workspace.get-user", "application/json", bytes.NewBuffer(j))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var result GetWorkspaceUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	var users []User

	for _, user := range result.Users {
		users = append(users, user)
	}

	ch <- users
}
