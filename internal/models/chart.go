package models

import "github.com/google/go-github/v45/github"

type Chart struct {
	github.RepositoryRelease
	Services []Release
}
