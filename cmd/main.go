package main

import (
	"context"
	"fmt"
	"os"

	awslogin "github.com/Photosynth-inc/aws-google-login"
	"github.com/urfave/cli/v3"
)

func handler(ctx context.Context, c *cli.Command) (err error) {
	g := awslogin.NewGoogleConfig(c.String("idp-id"), c.String("sp-id"))
	authnRequest, err := g.Login()
	if err != nil {
		return err
	}

	amz, err := awslogin.NewAWSConfig(authnRequest, c.Int("duration-seconds"))
	if err != nil {
		return err
	}
	principalArn, err := amz.GetPrincipalArn(c.String("role-arn"))
	if err != nil {
		return err
	}
	creds, err := amz.AssumeRole(ctx, c.String("role-arn"), principalArn)
	if err != nil {
		return err
	}

	awsCreds := &awslogin.AWSCredentials{
		Profile:     c.String("profile"),
		Credentials: creds,
	}
	return awsCreds.SaveTo(awslogin.AWSCredPath())
}

func main() {
	app := &cli.Command{
		Name:   "aws-google-login",
		Usage:  "Acquire temporary AWS credentials via Google SSO (SAML v2)",
		Action: handler,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "AWS Profile to use",
				Value:   "akerun",
			},
			&cli.IntFlag{
				Name:    "duration-seconds",
				Aliases: []string{"d"},
				Usage:   "Session Duration (in seconds)",
				Value:   3600,
			},
			&cli.StringFlag{
				Name:     "sp-id",
				Aliases:  []string{"s"},
				Usage:    "Service Provider ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "idp-id",
				Aliases:  []string{"i"},
				Usage:    "Identity Provider ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "role-arn",
				Aliases:  []string{"r"},
				Usage:    "AWS Role Arn for assuming to, ex: arn:aws:iam::123456789012:role/role-name",
				Required: true,
			},
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
