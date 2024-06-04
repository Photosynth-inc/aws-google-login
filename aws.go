package awslogin

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

const (
	XmlAttrGetRole            = "https://aws.amazon.com/SAML/Attributes/Role"
	XmlAttrGetRoleSessionName = "https://aws.amazon.com/SAML/Attributes/RoleSessionName"
	XmlAttrGetSessionDuration = "https://aws.amazon.com/SAML/Attributes/SessionDuration"
)

type AWS struct {
	AuthnRequest string
	Config       *AWSConfig
}

type Role struct {
	RoleArn      string `json:"role_arn"`
	PrincipalArn string `json:"principal_arn"`
}

func (r *Role) String() string {
	return fmt.Sprintf("RoleArn: %s, PrincipalArn: %s", r.RoleArn, r.PrincipalArn)
}

func NewAWSConfig(authnRequest string, config *AWSConfig) (*AWS, error) {
	if ok := IsValidSamlAssertion(authnRequest); !ok {
		return nil, fmt.Errorf("invalid SAML assertion")
	}

	return &AWS{
		AuthnRequest: authnRequest,
		Config:       config,
	}, nil
}

func (amz *AWS) parseRole(role string) (*Role, error) {
	items := strings.Split(role, ",")
	if len(items) != 2 {
		return nil, fmt.Errorf("invalid role string %v", role)
	}

	return &Role{
		RoleArn:      items[0],
		PrincipalArn: items[1],
	}, nil
}

func (amz *AWS) ParseRoles() ([]*Role, error) {
	resp := []*Role{}
	roleValues, err := GetAttributeValuesFromAssertion(amz.AuthnRequest, XmlAttrGetRole)
	if err != nil {
		return nil, err
	}

	for _, v := range roleValues {
		role, err := amz.parseRole(v)
		if err != nil {
			return nil, err
		}

		resp = append(resp, role)
	}

	return resp, nil
}

func (amz *AWS) ResolveRole(roleArn string) (*Role, error) {
	roles, err := amz.ParseRoles()
	if err != nil {
		return nil, err
	}

	for _, v := range roles {
		logger.Debug().Str("roleArn", roleArn).Str("v.RoleArn", v.RoleArn).Msg("role found")
		if roleArn == v.RoleArn {
			return v, nil
		}
	}
	return nil, fmt.Errorf("role is not configured for your user")
}

// AssumeRole is going to call sts.AssumeRoleWithSAMLInput to assume to a specific role
func (amz *AWS) AssumeRole(ctx context.Context, role *Role) (*types.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}
	svc := sts.NewFromConfig(cfg)
	input := &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: aws.Int32(int32(amz.Config.Google.Duration)),
		PrincipalArn:    aws.String(role.PrincipalArn),
		RoleArn:         aws.String(role.RoleArn),
		SAMLAssertion:   aws.String(amz.AuthnRequest),
	}

	result, err := svc.AssumeRoleWithSAML(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to assume role %v", err)
	}

	return result.Credentials, nil
}
