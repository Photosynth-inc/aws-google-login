package main

import (
	"context"
	"fmt"
	"os"

	awslogin "github.com/Photosynth-inc/aws-google-login"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

func handler(ctx context.Context, c *cli.Command) (err error) {
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

	authnRequest, err := g.Login()
	if err != nil {
		return err
	}

	amz, err := awslogin.NewAWSConfig(authnRequest, g.Google.Duration)
	if err != nil {
		return err
	}
	principalArn, err := amz.GetPrincipalArn(g.Google.RoleARN)
	if err != nil {
		return err
	}
	creds, err := amz.AssumeRole(ctx, g.Google.RoleARN, principalArn)
	if err != nil {
		return err
	}

	awsCreds := &awslogin.AWSCredentials{
		Profile:     g.Profile,
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
				Name:    "sp-id",
				Aliases: []string{"s"},
				Usage:   "Service Provider ID",
			},
			&cli.StringFlag{
				Name:    "idp-id",
				Aliases: []string{"i"},
				Usage:   "Identity Provider ID",
			},
			&cli.StringFlag{
				Name:    "role-arn",
				Aliases: []string{"r"},
				Usage:   "AWS Role Arn for assuming to, ex: arn:aws:iam::123456789012:role/role-name",
			},
			&cli.StringFlag{
				Name:       "log",
				Usage:      "change Log level, choose from: [trace | debug | info | warn | error | fatal | panic]",
				Persistent: true,
				Value:      "warn",
				Action: func(_ context.Context, cmd *cli.Command, flag string) error {
					if level, err := zerolog.ParseLevel(cmd.String("log")); err != nil {
						return err
					} else {
						zerolog.SetGlobalLevel(level)
						return nil
					}
				},
			},
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
