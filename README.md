# aws-google-login

This command-line tool allows you to acquire AWS temporary (STS) credentials using Google Apps as a federated (Single Sign-On, or SSO) provider. This project was inspired from [aws-google-auth](https://github.com/cevoaustralia/aws-google-auth) and the help of [playwright-go](https://github.com/mxschmitt/playwright-go) for the interactive Graphic User Interface (GUI).

This was hard-forked from [cucxabong/aws-google-login](https://github.com/cucxabong/aws-google-login).

## Installation

```bash
brew install Photosynth-inc/tap/aws-google-login
```

## Usage

```bash
$ make build
$ ./aws-google-login --help
NAME:
   aws-google-login - Acquire temporary AWS credentials via Google SSO (SAML v2)

USAGE:
   aws-google-login [global options] [command [command options]] [arguments...]

COMMANDS:
   config   Show current configuration
   cache    Manage application's cache
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --profile value, -p value           AWS Profile to use (default: "akerun")
   --duration-seconds value, -d value  Session Duration (in seconds) (default: 3600)
   --sp-id value, -s value             Service Provider ID (default value is in /Users/daikiwatanabe/.aws/config)
   --idp-id value, -i value            Identity Provider ID (default value is in /Users/daikiwatanabe/.aws/config)
   --role-arn value, -r value          AWS Role Arn for assuming to, ex: arn:aws:iam::123456789012:role/role-name
   --select-role-interactivelly, -l    choose AWS Role interactively. If set, 'role-arn' will be ignored (default: false)
   --browser-timeout value, -t value   browser timeout duration in seconds (default: 60)
   --log value                         change Log level, choose from: [trace | debug | info | warn | error | fatal | panic]
   --help, -h                          show help (default: false)
```
