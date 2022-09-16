package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/shurcooL/githubv4"

	"github.com/go-graphite/carbonapi/drone/pkg/github"
)

const CommenterLogin = "grafanabot"

func main() {
	grafanabotPat := getRequiredEnv("GRAFANABOT_PAT")

	repoOwner := getRequiredEnv("DRONE_REPO_OWNER")
	repoName := getRequiredEnv("DRONE_REPO_NAME")

	pullRequest, err := strconv.Atoi(getRequiredEnv("DRONE_PULL_REQUEST"))
	if err != nil {
		panic(err)
	}

	commentTypeIdentifier := flag.String("id", "", "String that identifies the comment type being submitted")
	commentBodyFilename := flag.String("bodyfile", "", "A file containing the comment body")
	flag.Parse()

	if *commentTypeIdentifier == "" {
		panic("Required argument: -i")
	}
	if *commentBodyFilename == "" {
		panic("Required argument: -b")
	}

	api := github.NewAPI(context.Background(), grafanabotPat)

	err = minimizeOutdatedComments(api, repoOwner, repoName, pullRequest, *commentTypeIdentifier)
	if err != nil {
		panic(err)
	}

	commentBody, err := ioutil.ReadFile(*commentBodyFilename)
	if err != nil {
		panic(err)
	}

	err = addComment(api, repoOwner, repoName, pullRequest, string(commentBody))
	if err != nil {
		panic(err)
	}
}

func getRequiredEnv(k string) string {
	v, p := os.LookupEnv(k)
	if !p {
		panic("Missing required env var: " + k)
	}

	return v
}

func minimizeOutdatedComments(api *github.API, repoOwner string, repoName string, pullRequestNo int, commentTypeIdentifier string) error {
	prComments, err := api.ListPullRequestComments(repoOwner, repoName, pullRequestNo)
	if err != nil {
		return err
	}

	for _, comment := range prComments {
		if comment.Author.Login == CommenterLogin && strings.Contains(comment.Body, commentTypeIdentifier) && !comment.IsMinimized {
			err := api.MinimizeComment(comment.ID, githubv4.ReportedContentClassifiersOutdated)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func addComment(api *github.API, repoOwner string, repoName string, pullRequestNo int, commentBody string) error {
	pullRequestNodeID, err := api.GetPullRequestNodeID(repoOwner, repoName, pullRequestNo)
	if err != nil {
		return err
	}

	_, err = api.AddComment(pullRequestNodeID, commentBody)
	if err != nil {
		return err
	}

	return nil
}
