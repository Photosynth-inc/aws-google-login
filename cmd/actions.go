package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		BrowserTimeout: c.Int("browser-timeout"),
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

		// 各カラムの最大幅を計算
		maxAliasLen := 0
		maxAccountIDLen := 0
		for _, r := range roles {
			if len(r.AccountAlias) > maxAliasLen {
				maxAliasLen = len(r.AccountAlias)
			}
			if len(r.AccountID()) > maxAccountIDLen {
				maxAccountIDLen = len(r.AccountID())
			}
		}

		type displayRole struct {
			Alias     string
			AccountID string
			RoleName  string
			Original  *awslogin.Role
		}

		displayRoles := make([]displayRole, len(roles))
		for i, r := range roles {
			displayRoles[i] = displayRole{
				Alias:     fmt.Sprintf("%-*s", maxAliasLen, r.AccountAlias),
				AccountID: fmt.Sprintf("%-*s", maxAccountIDLen, r.AccountID()),
				RoleName:  r.RoleName(),
				Original:  r,
			}
		}

		prompt := promptui.Select{
			Label: "Select AWS Role",
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   "▶ {{ .Alias | cyan }} ({{ .AccountID | red }}) # {{ .RoleName | yellow }}",
				Inactive: "  {{ .Alias | cyan }} ({{ .AccountID | red }}) # {{ .RoleName | yellow }}",
				Selected: "▶ {{ .Alias | cyan }} ({{ .AccountID | red }}) # {{ .RoleName | yellow }}",
			},
			Items: displayRoles,
			Size:  10,
			Searcher: func(input string, index int) bool {
				r := displayRoles[index]
				return strings.Contains(r.Alias, input) ||
					strings.Contains(r.AccountID, input) ||
					strings.Contains(r.RoleName, input)
			},
		}
		index, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed %v", err)
		}
		role = displayRoles[index].Original
	} else {
		role, err = amz.ResolveRole(g.Google.RoleARN)
		if err != nil {
			return err
		}
	}

	creds, err := amz.AssumeRole(ctx, role, &awslogin.AssumeRoleOptions{
		ReadOnly:        c.Bool("read-only"),
		BillingReadOnly: c.Bool("billing-read-only"),
	})
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
		return nil
	}
	if c.Bool("credentials_path") {
		fmt.Println(awslogin.AWSCredPath())
		return nil
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
		fmt.Printf("Are you sure to clear cache? [y/N] ")
		var ans string
		_, err := fmt.Fscanln(os.Stdin, &ans)
		if err != nil {
			panic(err)
		}
		if ans != "y" {
			fmt.Println("Aborted")
			return nil
		}

		if err := awslogin.DeleteBrowserCache(); err != nil {
			return err
		}
		fmt.Println("Cache has been cleared")
		return nil
	}

	fmt.Println(awslogin.ConfigDirRoot())
	return nil
}

func handleStatus(ctx context.Context, c *cli.Command) error {
	profile := c.String("profile")
	status, err := awslogin.LoadSessionStatus(awslogin.AWSCredPath(), profile)
	if err != nil {
		fmt.Printf("No valid session found for profile %q\n", profile)
		if c.Bool("quiet") {
			os.Exit(1)
		}
		return nil
	}

	if c.Bool("quiet") {
		if status.IsValid {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if status.IsValid {
		hours := int(status.TimeLeft.Hours())
		minutes := int(status.TimeLeft.Minutes()) % 60
		seconds := int(status.TimeLeft.Seconds()) % 60
		fmt.Printf("Profile:    %s\n", status.Profile)
		fmt.Printf("Status:     valid\n")
		fmt.Printf("Expires:    %s\n", status.Expiration.Local().Format("2006-01-02 15:04:05"))
		fmt.Printf("Time left:  %dh %dm %ds\n", hours, minutes, seconds)
	} else {
		fmt.Printf("Profile:    %s\n", status.Profile)
		fmt.Printf("Status:     expired\n")
		fmt.Printf("Expired at: %s\n", status.Expiration.Local().Format("2006-01-02 15:04:05"))
	}

	return nil
}
