package main

import (
	"context"
	"os"

	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

const (
	Nick = "notetaker"
)

var (
	Username    = os.Getenv("SASL_USER")
	Password    = os.Getenv("SASL_PASSWORD")
	Server      = os.Getenv("SERVER")
	VerifyTLS   = os.Getenv("VERIFY_TLS") == "true"
	GithubToken = os.Getenv("GITHUB_TOKEN")
)

func must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}

func githubClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc)
}

func main() {
	c, err := New(Username, Password, Server, VerifyTLS, githubClient(GithubToken))
	if err != nil {
		panic(err)
	}

	c.bottom.Client.Connect()
}
