package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	awslogin "github.com/Photosynth-inc/aws-google-login"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/urfave/cli/v3"
)

func handler(_ context.Context, c *cli.Command) (err error) {
	g := awslogin.NewGoogleConfig(c.String("idp-id"), c.String("sp-id"))
	assertion, err := g.Login()
	if err != nil {
		return err
	}

	if c.Bool("get-saml-assertion") {
		_, err := fmt.Println(assertion)
		return err
	}

	amz, err := awslogin.NewAmazonConfig(assertion, c.Int("duration-seconds"))
	if err != nil {
		return err
	}

	s, err := AssumeRole(amz, c.String("role-arn"))
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

func AssumeRole(amz *awslogin.Amazon, roleArn string) (*types.Credentials, error) {
	var principalArn string
	roles, err := amz.ParseRoles()
	if err != nil {
		return nil, err
	}

	for _, v := range roles {
		if roleArn == v.RoleArn {
			principalArn = v.PrincipalArn
			break
		}
	}

	if principalArn == "" {
		fmt.Println(roleArn, roles)
		return nil, fmt.Errorf("role is not configured for your user")
	}

	return amz.AssumeRole(context.TODO(), roleArn, principalArn)
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
		&cli.BoolFlag{
			Name:    "get-saml-assertion",
			Aliases: []string{"l"},
			Usage:   "Getting SAML assertion XML",
			Value:   false,
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
