package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ReleasesInfo struct {
	Id      uint   `json:"id"`
	TagName string `json:"tag_name"`
}

type ReleaseInfoer interface {
	GetLatestReleaseTag(string) (string, error)
}

type GithubReleaseInfoer struct{}

func (gh GithubReleaseInfoer) GetLatestReleaseTag(repo string) (string, error) {
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)
	response, err := http.Get(apiUrl)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	releases := []ReleasesInfo{}

	if err := json.Unmarshal(body, &releases); err != nil {
		return "", err
	}

	tag := releases[0].TagName

	return tag, nil
}

// Function to get the message to display to the end user.
func getReleaseTagMessage(ri ReleaseInfoer, repo string) (string, error) {
	tag, err := ri.GetLatestReleaseTag(repo)
	if err != nil {
		return "", fmt.Errorf("Error querying GitHub API: %s", err)
	}

	return fmt.Sprintf("The latest release is %s", tag), nil
}

func main() {
	gh := GithubReleaseInfoer{}
	msg, err := getReleaseTagMessage(gh, "docker/machine")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(msg)
}
