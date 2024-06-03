# aws-google-login

This command-line tool allows you to acquire AWS temporary (STS) credentials using Google Apps as a federated (Single Sign-On, or SSO) provider. This project was inspired from [aws-google-auth](https://github.com/cevoaustralia/aws-google-auth)
 and the help of [playwright-go](https://github.com/mxschmitt/playwright-go) for the interactive Graphic User Interface (GUI)

## Usage

```bash
$ make build
$ ./aws-google-login --help
NAME:
   aws-google-login - Acquire temporary AWS credentials via Google SSO (SAML v2)

USAGE:
   aws-google-login [global options] [command [command options]] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --profile value, -p value           AWS Profile to use (default: "akerun")
   --duration-seconds value, -d value  Session Duration (in seconds) (default: 3600)
   --sp-id value, -s value             Service Provider ID
   --idp-id value, -i value            Identity Provider ID
   --role-arn value, -r value          AWS Role Arn for assuming to, ex: arn:aws:iam::123456789012:role/role-name
   --log value                         change Log level, choose from: [trace | debug | info | warn | error | fatal | panic] (default: "warn")
   --help, -h                          show help (default: false)
```
