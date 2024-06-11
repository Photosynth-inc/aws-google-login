package awslogin

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"gopkg.in/ini.v1"
)

// AWSConfig reflects values in the AWS CLI config file (mainly as `~/.aws/config`)
type AWSConfig struct {
	Profile string
	Region  string
	Google  AWSConfig_GoogleConfig
}

func (c *AWSConfig) String() string {
	return fmt.Sprintf(
		"Profile: %s\nRegion: %s\nGoogle_Configuration:\n  Ask_Role: %v\n  Keyring: %v\n  Duration: %d\n  Google_IDP_ID: %s\n  Google_SP_ID: %s\n  U2F_Disabled: %v\n  Google_Username: %s\n  BG_Response: %s\n  Role_ARN: %s",
		c.Profile,
		c.Region,
		c.Google.AskRole,
		c.Google.Keyring,
		c.Google.Duration,
		c.Google.GoogleIDPID,
		c.Google.GoogleSPID,
		c.Google.U2FDisabled,
		c.Google.GoogleUserName,
		c.Google.BGResponse,
		c.Google.RoleARN,
	)
}

type AWSConfig_GoogleConfig struct {
	AskRole        bool
	Keyring        bool
	Duration       int64
	GoogleIDPID    string
	GoogleSPID     string
	U2FDisabled    bool
	GoogleUserName string
	BGResponse     string
	RoleARN        string
}

func LoadConfig(path, profile string) (*AWSConfig, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	section, err := cfg.GetSection("profile " + profile)
	if err != nil {
		return nil, err
	}
	return &AWSConfig{
		Profile: profile,
		Region:  section.Key("region").String(),
		Google: AWSConfig_GoogleConfig{
			AskRole:        section.Key("google_config.ask_role").MustBool(true),
			Keyring:        section.Key("google_config.keyring").MustBool(false),
			Duration:       section.Key("google_config.duration").MustInt64(3600),
			GoogleIDPID:    section.Key("google_config.google_idp_id").String(),
			GoogleSPID:     section.Key("google_config.google_sp_id").String(),
			U2FDisabled:    section.Key("google_config.u2f_disabled").MustBool(false),
			GoogleUserName: section.Key("google_config.google_username").String(),
			BGResponse:     section.Key("google_config.bg_response").String(),
			RoleARN:        section.Key("google_config.role_arn").String(),
		},
	}, nil
}

// AWSCredentials reflects values in the AWS CLI credentials file (mainly as `~/.aws/credentials`)
type AWSCredentials struct {
	Profile string
	*types.Credentials
}

func (cred *AWSCredentials) SaveTo(path string) error {
	cfg, err := ini.Load(path)
	if err != nil {
		return err
	}

	section := cfg.Section(cred.Profile)
	section.Key("aws_access_key_id").SetValue(*cred.AccessKeyId)
	section.Key("aws_secret_access_key").SetValue(*cred.SecretAccessKey)
	section.Key("aws_session_token").SetValue(*cred.SessionToken)
	section.Key("aws_session_expiration").SetValue(cred.Expiration.Format("2006-01-02T15:04:05Z"))

	return cfg.SaveTo(path)
}
