package awslogin

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"golang.org/x/sync/errgroup"
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
	AccountAlias string `json:"account_alias"`
}

func (r *Role) String() string {
	if r.AccountAlias != "" {
		return fmt.Sprintf("%s (%s)", r.AccountAlias, r.AccountID())
	} else {
		return r.RoleArn
	}
}

func (r *Role) AccountID() string {
	items := strings.Split(r.RoleArn, ":") // arn:aws:iam::123456789012:role/role-name -> [arn, aws, iam, 123456789012, role, role-name]
	if len(items) < 5 {
		return "unknown"
	}
	return items[4]
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

func (amz *AWS) createConfig(creds *types.Credentials) aws.Config {
	credProvider := credentials.NewStaticCredentialsProvider(
		*creds.AccessKeyId,
		*creds.SecretAccessKey,
		*creds.SessionToken,
	)

	return aws.Config{
		Credentials: credProvider,
		Region:      amz.Config.Region,
	}
}

// resolveAlias resolves the account alias with the given role.
// This assumes that there is only one account alias for the given role.
func (amz *AWS) resolveAlias(ctx context.Context, role *Role) (*Role, error) {
	creds, err := amz.AssumeRole(ctx, role)
	if err != nil {
		return nil, err
	}

	out, err := iam.
		NewFromConfig(amz.createConfig(creds)).
		ListAccountAliases(ctx, &iam.ListAccountAliasesInput{})

	if err != nil {
		logger.Debug().Err(err).Msg("ListAccountAliases failed, fallback to account ID")
		role.AccountAlias = role.AccountID()
	} else if len(out.AccountAliases) == 0 {
		logger.Debug().Msg("no account alias found, fallback to account ID")
		role.AccountAlias = role.AccountID()
	} else {
		role.AccountAlias = out.AccountAliases[0]
	}

	return role, nil
}

func (amz *AWS) ResolveAliases(ctx context.Context) ([]*Role, error) {
	roles, err := amz.ParseRoles()
	if err != nil {
		return nil, err
	}

	eg := errgroup.Group{}

	// Resolve all aliases
	for i, role := range roles {
		i, role := i, role
		eg.Go(func() error {
			resolved, err := amz.resolveAlias(ctx, role)
			if err != nil {
				return err
			}

			logger.Debug().Str("alias", fmt.Sprintf("%+v", resolved.AccountAlias)).Msg("account alias found")
			roles[i] = resolved
			return nil
		})
	}
	err = eg.Wait()

	// Sort the roles by account alias alphabetically
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].AccountAlias < roles[j].AccountAlias
	})
	return roles, err
}
