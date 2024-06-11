package main

import (
	"context"
	"fmt"

	awslogin "github.com/Photosynth-inc/aws-google-login"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

func handleMain(ctx context.Context, c *cli.Command) (err error) {
	g, err := awslogin.LoadConfig(awslogin.AWSConfigPath(), c.String("profile"))
	if err != nil {
		return err
	}

	// Override the configuration if the flags are set
	{
		if c.String("sp-id") != "" {
			g.Google.GoogleSPID = c.String("sp-id")
		}
		if c.String("idp-id") != "" {
			g.Google.GoogleIDPID = c.String("idp-id")
		}
		if c.String("role-arn") != "" {
			g.Google.RoleARN = c.String("role-arn")
		}
		if c.Int("duration-seconds") != 0 {
			g.Google.Duration = c.Int("duration-seconds")
		}
	}

	authnRequest, err := g.Login(&awslogin.LoginOptions{
		Verbose:        zerolog.GlobalLevel() < zerolog.WarnLevel,
		BrowserTimeout: c.Float("browser-timeout"),
	})
	if err != nil {
		return err
	}

	amz, err := awslogin.NewAWSConfig(authnRequest, g)
	if err != nil {
		return err
	}

	var role *awslogin.Role
	if c.Bool("select-role-interactivelly") {
		roles, err := amz.ResolveAliases(ctx)
		if err != nil {
			return err
		}
		prompt := promptui.Select{
			Label: "Select AWS Role",
			Items: roles,
			Size:  10,
		}
		index, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed %v", err)
		}
		role = roles[index]
	} else {
		role, err = amz.ResolveRole(g.Google.RoleARN)
		if err != nil {
			return err
		}
	}

	creds, err := amz.AssumeRole(ctx, role)
	if err != nil {
		return err
	}

	awsCreds := &awslogin.AWSCredentials{
		Profile:     g.Profile,
		Credentials: creds,
	}

	if err := awsCreds.SaveTo(awslogin.AWSCredPath()); err != nil {
		return err
	}
	fmt.Println("Temporary AWS credentials have been saved to", awslogin.AWSCredPath())
	return nil
}

func handleConfig(ctx context.Context, c *cli.Command) error {
	if c.Bool("config_path") {
		fmt.Println(awslogin.AWSConfigPath())
	}
	if c.Bool("credentials_path") {
		fmt.Println(awslogin.AWSCredPath())
	}

	cfg, err := awslogin.LoadConfig(awslogin.AWSConfigPath(), c.String("profile"))
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", cfg)
	return nil
}

func handleCache(ctx context.Context, c *cli.Command) error {
	if c.Bool("clear") {
		if err := awslogin.DeleteBrowserCache(); err != nil {
			return err
		}
		fmt.Println("Cache has been cleared")
	}
	fmt.Println(awslogin.ConfigDirRoot())
	return nil
}
