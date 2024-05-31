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

type AWS struct {
	AuthnRequest    string
	SessionDuration int64
}

type Role struct {
	RoleArn      string `json:"role_arn"`
	PrincipalArn string `json:"principal_arn"`
}

func (r *Role) String() string {
	return fmt.Sprintf("RoleArn: %s, PrincipalArn: %s", r.RoleArn, r.PrincipalArn)
}

func NewAWSConfig(authnRequest string, sessionDuration int64) (*AWS, error) {
	if ok := IsValidSamlAssertion(authnRequest); !ok {
		return nil, fmt.Errorf("invalid SAML assertion")
	}

	return &AWS{
		AuthnRequest:    authnRequest,
		SessionDuration: sessionDuration,
	}, nil
}

func (amz *AWS) GetPrincipalArn(roleArn string) (string, error) {
	roles, err := amz.ParseRoles()
	if err != nil {
		return "", err
	}

	for _, v := range roles {
		if roleArn == v.RoleArn {
			return v.PrincipalArn, nil
		}
	}

	return "", fmt.Errorf("role is not configured for your user")
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
	roleValues, err := GetAttributeValuesFromAssertion(amz.AuthnRequest, amz.GetRoleAttrName())
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

// GetRoleAttrName return XML attribute name for Role property
func (*AWS) GetRoleAttrName() string {
	return "https://aws.amazon.com/SAML/Attributes/Role"
}

// GetRoleSessionNameAttrName return XML attribute name for RoleSessionName property
func (*AWS) GetRoleSessionNameAttrName() string {
	return "https://aws.amazon.com/SAML/Attributes/RoleSessionName"
}

// GetSessionDurationAttrName return XML attribute name for SessionDuration property
func (*AWS) GetSessionDurationAttrName() string {
	return "https://aws.amazon.com/SAML/Attributes/SessionDuration"
}

// AssumeRole is going to call sts.AssumeRoleWithSAMLInput to assume to a specific role
func (amz *AWS) AssumeRole(ctx context.Context, roleArn, principalArn string) (*types.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}
	svc := sts.NewFromConfig(cfg)
	input := &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: aws.Int32(int32(amz.SessionDuration)),
		PrincipalArn:    aws.String(principalArn),
		RoleArn:         aws.String(roleArn),
		SAMLAssertion:   aws.String(amz.AuthnRequest),
	}

	result, err := svc.AssumeRoleWithSAML(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to assume role %v", err)
	}

	return result.Credentials, nil
}
