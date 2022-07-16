package util

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

func Select(msg string, elems []string) (int, string, error) {
	fmt.Println()
	prompt2 := promptui.Select{
		Label: msg,
		Items: elems,
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
