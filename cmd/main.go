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

type Options struct {
	ServiceProviderID  string
	IdentityProviderID string
	DurationSeconds    int64
	RoleName           string
	AccountIDs         []string
	SamlAssertion      bool
}

type CredentialsData struct {
	*types.Credentials
	AccountId string
	RoleArn   string
}

func NewOptions(c *cli.Command) *Options {
	return &Options{
		ServiceProviderID:  c.String("sp-id"),
		IdentityProviderID: c.String("idp-id"),
		DurationSeconds:    c.Int("duration-seconds"),
		RoleName:           c.String("role-name"),
		AccountIDs:         c.StringSlice("account-ids"),
		SamlAssertion:      c.Bool("get-saml-assertion"),
	}
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

func handler(_ context.Context, c *cli.Command) error {
	var assertion string
	var err error
	opt := NewOptions(c)

	g := awslogin.NewGoogleConfig(opt.IdentityProviderID, opt.ServiceProviderID)
	assertion, err = g.Login()
	if err != nil {
		return err
	}

	if opt.SamlAssertion {
		_, err := fmt.Println(assertion)
		return err
	}

	amz, err := awslogin.NewAmazonConfig(assertion, int64(opt.DurationSeconds))
	if err != nil {
		return err
	}

	creds := make([]CredentialsData, len(opt.AccountIDs))

	for idx, accountID := range opt.AccountIDs {
		roleArn := GetRoleArn(accountID, opt.RoleName)
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

	JSONWrite(os.Stdout, creds)
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
			Value:   43200,
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
			Name:  "saml-assertion",
			Usage: "Getting SAML assertion XML",
			Value: false,
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
