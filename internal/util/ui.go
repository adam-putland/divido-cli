package util

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

type Options []string

func (options Options) WithBackOption() Options {
	return append(options, "Back")
}

func Select(msg string, elems Options) (int, string, error) {
	fmt.Println()
	prompt2 := promptui.Select{

		Label: msg,
		Items: elems,
	}
	return prompt2.Run()
}

func SelectWithSearch(msg string, elems Options, searcher func(input string, index int) bool) (int, string, error) {
	fmt.Println()
	prompt2 := promptui.Select{
		Label:    msg,
		Items:    elems,
		Searcher: searcher,
	}
	return prompt2.Run()
}

func SelectWithAdd(msg string, elems Options) (int, string, error) {
	fmt.Println()
	fmt.Println()
	prompt2 := promptui.SelectWithAdd{
		Label:    msg,
		Items:    elems,
		AddLabel: "Other",
	}
	return prompt2.Run()
}

func Prompt(msg string) (string, error) {
	prompt := promptui.Prompt{
		Label:    msg,
		Validate: func(input string) error { return nil },
	}

	return prompt.Run()
}

func PromptWithDefault(msg, defaultMsg string) (string, error) {
	prompt := promptui.Prompt{
		Label:    msg,
		Validate: func(input string) error { return nil },
		Default:  defaultMsg,
	}

	return prompt.Run()
}
