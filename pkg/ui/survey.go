package ui

import (
	"github.com/AlecAivazis/survey/v2"
)

func MultiSelectPrompt(message string, options, defaultValues []string) ([]string, error) {
	var selected []string

	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
		Default: defaultValues,
	}

	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return nil, err
	}

	return selected, nil
}
