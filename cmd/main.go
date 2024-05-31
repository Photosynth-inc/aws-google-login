package main

import (
	"context"
	"encoding/json"
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

	s, err := amz.AssumeRole(ctx, c.String("role-arn"), principalArn)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(s)
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))
	return nil
}

func main() {
	app := &cli.Command{
		Name:   "aws-google-login",
		Usage:  "Acquire temporary AWS credentials via Google SSO (SAML v2)",
		Action: handler,
	}
	app.Flags = []cli.Flag{
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
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
