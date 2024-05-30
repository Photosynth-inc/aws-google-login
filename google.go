package awslogin

import (
	"fmt"
	"log"
	"net/url"

	"github.com/playwright-community/playwright-go"
)

type Google struct {
	IdpID string
	SpID  string
}

func (g *Google) LoginURL() string {
	return fmt.Sprintf("https://accounts.google.com/o/saml2/initsso?idpid=%s&spid=%s&forceauthn=false",
		g.IdpID, g.SpID)
}

func (g *Google) WaitURL() string {
	return "https://signin.aws.amazon.com/saml"
}

func NewGoogleConfig(idpID, spID string) *Google {
	return &Google{
		IdpID: idpID,
		SpID:  spID,
	}
}

func (g *Google) Login() (string, error) {
	if err := playwright.Install(); err != nil {
		return "", fmt.Errorf("could not install playwright: %v", err)
	}

	SAMLResponse := ""

	pw, err := playwright.Run(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		return SAMLResponse, fmt.Errorf("unable to run playwright %v", err)
	}

	browser, err := pw.Chromium.LaunchPersistentContext(ConfigEntry("browser"), playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return SAMLResponse, fmt.Errorf("could not launch a browser %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		return SAMLResponse, fmt.Errorf("could not create page: %v", err)
	}

	if _, err := page.Goto(g.LoginURL()); err != nil {
		return SAMLResponse, fmt.Errorf("could not goto: %v", err)
	}

	r, err := page.ExpectRequest(g.WaitURL(), func() error {
		return nil
	})
	if err != nil {
		return SAMLResponse, fmt.Errorf("could not wait for request: %v", err)
	}
	data, err := r.PostData()
	if err != nil {
		return SAMLResponse, fmt.Errorf("can not get PostData %v", err)
	}

	values, err := url.ParseQuery(data)
	if err != nil {
		return SAMLResponse, fmt.Errorf("unable to parse PostData %v", err)
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	return values.Get("SAMLResponse"), nil
}
