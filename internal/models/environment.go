package models

import "github.com/google/go-github/v45/github"

type Environment struct {
	github.RepositoryRelease
	Chart     Chart
	Overrides []Release
}
