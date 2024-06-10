package awslogin

import (
	"fmt"
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

// Login invokes the Playwright browser to login to Google,
// and returns the `AuthnRequest` (SAMLResponse) captured from the browser request.
func (g *Google) Login() (resp string, err error) {
	if err := playwright.Install(); err != nil {
		return "", fmt.Errorf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		return "", fmt.Errorf("unable to run playwright %v", err)
	}

	browser, err := pw.Chromium.LaunchPersistentContext(ConfigEntry("browser"), playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return "", fmt.Errorf("could not launch a browser %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		return "", fmt.Errorf("could not create page: %v", err)
	}

	page.OnRequest(func(req playwright.Request) {
		if req.URL() == g.WaitURL() {
			fmt.Println("Request received, processing...")
			data, _ := req.PostData()
			values, _ := url.ParseQuery(data)
			resp = values.Get("SAMLResponse")
		}
	})

	fmt.Println("Please login to your Google account and press any key to continue...")
	if _, err := page.Goto(g.LoginURL()); err != nil {
		return "", fmt.Errorf("could not goto: %v", err)
	}
	page.WaitForURL(g.WaitURL())

	if err = page.Close(); err != nil {
		return "", fmt.Errorf("could not close page: %v", err)
	}
	if err = browser.Close(); err != nil {
		return "", fmt.Errorf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		return "", fmt.Errorf("could not stop Playwright: %v", err)
	}
	return resp, nil
}
