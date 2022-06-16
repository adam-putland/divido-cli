package internal

import "github.com/manifoldco/promptui"

func Select(msg string, elems []string) (int, string, error) {
	prompt2 := promptui.Select{
		Label: msg,
		Items: elems,
	}
	return prompt2.Run()
}

func Prompt(msg string, defaultMsg ...string) (string, error) {
	prompt := promptui.Prompt{
		Label:    msg,
		Validate: func(input string) error { return nil },
		Default:  defaultMsg[0],
	}

	return prompt.Run()
}
