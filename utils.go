package awslogin

import (
	"os"
	"path/filepath"
	"time"

	"github.com/RobotsAndPencils/go-saml"
)

func IsValidSamlAssertion(assertion string) bool {
	if len(assertion) == 0 {
		return false
	}

	parsedSaml, err := saml.ParseEncodedResponse(assertion)
	if err != nil {
		return false
	}

	notBefore, err := time.Parse(time.RFC3339Nano, parsedSaml.Assertion.Conditions.NotBefore)
	if err != nil {
		return false
	}
	notOnOrAfter, err := time.Parse(time.RFC3339Nano, parsedSaml.Assertion.Conditions.NotOnOrAfter)
	if err != nil {
		return false
	}
	now := time.Now()

	if now.Before(notBefore) || now.After(notOnOrAfter) || now.Equal(notOnOrAfter) {
		return false
	}

	return true
}

// GetAttributeValuesFromAssertion parse SAML Assertion in form of XML document
// to return a list of attribute values from attribute name
func GetAttributeValuesFromAssertion(assertion, attributeName string) ([]string, error) {
	parsedSaml, err := saml.ParseEncodedResponse(assertion)
	if err != nil {
		return nil, err
	}

	return parsedSaml.GetAttributeValues(attributeName), nil
}

func ConfigDirRoot() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(configDir, "aws-google-login")
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	return dir
}

func ConfigEntry(name string) string {
	return filepath.Join(ConfigDirRoot(), name)
}
