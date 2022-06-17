package models

import (
	"github.com/google/go-github/v45/github"
)

type Service struct {
	github.RepositoryRelease
}