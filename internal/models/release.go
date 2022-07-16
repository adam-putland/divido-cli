package models

import "fmt"

type Releases []*Release
type Versions []string

type Release struct {
	Name      string
	Version   string
	Changelog string
	URL       string
}

func (releases Releases) String() string {
	var str string
	for _, r := range releases {
		str += fmt.Sprintf("%s\n", r.Version)
	}
	return str
}

func (releases Releases) Versions() Versions {
	versions := make([]string, 0, len(releases))
	for _, r := range releases {
		versions = append(versions, r.Version)
	}
	return versions
}

func (release Release) String() string {
	return fmt.Sprintf(" Name: %s\n latest version: %s\n URL: %s", release.Name, release.Version, release.URL)
}

func (versions *Versions) Remove(versionIndex int) {
	copy((*versions)[versionIndex:], (*versions)[versionIndex+1:]) // shift valuesafter the indexwith a factor of 1
	(*versions)[len(*versions)-1] = ""                             // remove element
	*versions = (*versions)[:len(*versions)-1]
}
