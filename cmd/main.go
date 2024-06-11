package main

import (
	"context"
	"fmt"
	"os"

	awslogin "github.com/Photosynth-inc/aws-google-login"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:   "aws-google-login",
		Usage:  "Acquire temporary AWS credentials via Google SSO (SAML v2)",
		Action: handleMain,
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
				Usage:   fmt.Sprintf("Service Provider ID (default value is in %s)", awslogin.AWSConfigPath()),
			},
			&cli.StringFlag{
				Name:    "idp-id",
				Aliases: []string{"i"},
				Usage:   fmt.Sprintf("Identity Provider ID (default value is in %s)", awslogin.AWSConfigPath()),
			},
			&cli.StringFlag{
				Name:    "role-arn",
				Aliases: []string{"r"},
				Usage:   "AWS Role Arn for assuming to, ex: arn:aws:iam::123456789012:role/role-name",
			},
			&cli.BoolFlag{
				Name:    "select-role-interactivelly",
				Aliases: []string{"l"},
				Usage:   "choose AWS Role interactively. If set, 'role-arn' will be ignored",
				Value:   false,
			},
			&cli.FloatFlag{
				Name:    "browser-timeout",
				Aliases: []string{"t"},
				Usage:   "browser timeout duration in seconds",
				Value:   60.0,
			},
			&cli.StringFlag{
				Name:       "log",
				Usage:      "change Log level, choose from: [trace | debug | info | warn | error | fatal | panic]",
				Persistent: true,
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
		Commands: []*cli.Command{
			{
				Name:   "config",
				Usage:  "Show current configuration",
				Action: handleConfig,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "config_path",
						Usage: "Show the aws configuration path",
					},
					&cli.BoolFlag{
						Name:  "credentials_path",
						Usage: "Show the aws credentials path",
					},
				},
			},
			{
				Name:   "cache",
				Usage:  "Manage application's cache",
				Action: handleCache,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "clear",
						Usage: "Clear the browser cache (this is not reversible!)",
					},
				},
			},
		},
	}

	// set default / weirdly --log flag does not work if not set
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
