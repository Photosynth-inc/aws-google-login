package main

import (
	awslogin "github.com/cucxabong/aws-google-login"
	"github.com/manifoldco/promptui"
)

func interactiveAssumeRole(amz *awslogin.Amazon, export bool) error {

	roles, err := amz.ParseRoles()
	if err != nil {
		return err
	}
	if len(roles) == 1 {
		return assumeSingleRoleHandler(amz, roles[0].RoleArn, export)
	}

	templates := promptui.SelectTemplates{
		Active:   `🔐 {{ .RoleArn | cyan | bold }}`,
		Inactive: `   {{ .RoleArn | cyan }}`,
		Selected: `{{ "✔" | green | bold }} {{ "Assuming to" | bold }}: {{ .RoleArn | cyan }}`,
	}

	list := promptui.Select{
		Label:     "Select a role",
		Items:     roles,
		Templates: &templates,
		Size:      len(roles),
	}

	_, selected, err := list.Run()
	if err != nil {
		panic(err)
	}

	return assumeSingleRoleHandler(amz, selected, export)
}
