package models

import "fmt"

type Releases []*Release

func (releases Releases) String() string {
	var str string
	for _, r := range releases {
		str += fmt.Sprintf("%s\n", r.Version)
	}
	return str
}

func (releases Releases) Versions() []string {
	versions := make([]string, 0, len(releases))
	for _, r := range releases {
		versions = append(versions, r.Version)
	}
	return versions
}

type Release struct {
	Name      string
	Version   string
	Changelog string
	URL       string
}
