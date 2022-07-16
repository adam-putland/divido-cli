package models

import "fmt"

type Comparer struct {
	Changelog string
	Diff      string
}

func (c Comparer) String() string {
	//TODO implement me
	return fmt.Sprintf(" Changelog: \n %s \n Diff: \n %s", c.Changelog, c.Diff)
}
