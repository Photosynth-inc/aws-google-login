package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	awslogin "github.com/Photosynth-inc/aws-google-login"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/urfave/cli/v3"
)

type CredentialsData struct {
	*types.Credentials
	AccountId string
	RoleArn   string
}

func GetRoleArn(accountID string, roleName string) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName)
}

func JSONWrite(w io.Writer, data []CredentialsData) error {
	for _, item := range data {
		jsonData, err := json.Marshal(item)
		if err != nil {
			return err
		}

		if _, err = fmt.Fprintln(w, string(jsonData)); err != nil {
			return err
		}
	}
	return nil
}

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

	creds := make([]CredentialsData, len(c.StringSlice("account-ids")))

	for idx, accountID := range c.StringSlice("account-ids") {
		roleArn := GetRoleArn(accountID, c.String("role-name"))
		s, err := AssumeRole(amz, roleArn)
		if err != nil {
			return err
		}

		creds[idx] = CredentialsData{
			Credentials: s,
			AccountId:   accountID,
			RoleArn:     roleArn,
		}
	}

	return JSONWrite(os.Stdout, creds)
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
			Name:     "role-name",
			Aliases:  []string{"r"},
			Usage:    "AWS Role Arn for assuming to",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:     "account-ids",
			Aliases:  []string{"a"},
			Usage:    "AWS Account ID (can be specified multiple times)",
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
