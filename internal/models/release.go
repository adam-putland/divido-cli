package models

import (
	"fmt"
	"time"
)

type Releases []*Release
type Versions []string

type Release struct {
	Name      string
	Version   string
	Changelog string
	URL       string
	Date      time.Time
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

func (releases Releases) GetReleaseByVersion(version string) *Release {
	for _, r := range releases {
		if r.Version == version {
			return r
		}
	}
	return nil
}

func (release Release) String() string {
	return fmt.Sprintf(" Name: %s\n latest version: %s\n URL: %s\n", release.Name, release.Version, release.URL)
}

func (versions *Versions) Remove(index int) {
	copy((*versions)[index:], (*versions)[index+1:]) // shift valuesafter the indexwith a factor of 1
	(*versions)[len(*versions)-1] = ""               // remove element
	*versions = (*versions)[:len(*versions)-1]
}
