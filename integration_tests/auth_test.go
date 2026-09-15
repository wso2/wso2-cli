package integrationtests_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/wso2/integration-platform-tools/internal/auth"
)

func TestLogin(t *testing.T) {
	if TEST_USER_NAME == "" || TEST_USER_PASS == "" {
		t.Fatal("CHOREO_CLI_TEST_USER_NAME and CHOREO_CLI_TEST_USER_PASS environment variables are required")
	}

	log.Printf("=== Starting Login Test ===")
	log.Printf("Test environment: %s", ENV)
	log.Printf("Test project name: %s", TestProjectName)

	// Create Chrome context with CI-friendly options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("headless", true), // Force headless mode for CI
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := os.MkdirAll("recording", 0755); err != nil {
		t.Fatal("failed to create recording directory:", err)
	}

	var frameNum int
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*page.EventScreencastFrame); ok {
			frameNum++
			filename := fmt.Sprintf("recording/frame-%04d.jpeg", frameNum)
			data, err := base64.StdEncoding.DecodeString(e.Data)
			if err != nil {
				log.Printf("could not decode frame: %v", err)
				return
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				log.Printf("could not write screenshot to %s: %v", filename, err)
			}
			go func() {
				ackErr := chromedp.Run(ctx, page.ScreencastFrameAck(e.SessionID))
				if ackErr != nil {
					if !strings.Contains(ackErr.Error(), "context canceled") {
						log.Printf("could not ack frame: %v", ackErr)
					}
				}
			}()
		}
	})

	log.Printf("Starting auth flow...")
	port, link, err := auth.StartAuthFlow()
	if err != nil {
		log.Printf("Failed to start auth flow: %v", err)
		t.Fatal(err)
	}
	log.Printf("Auth flow started successfully on port %d", port)
	log.Printf("Auth URL: %s", link)

	log.Printf("Starting browser automation...")
	go func() {
		log.Printf("Navigating to auth URL...")
		err = chromedp.Run(ctx, chromedp.Tasks{
			page.StartScreencast().WithFormat("jpeg").WithQuality(80),
			// Initial page load with timeout
			chromedp.Navigate(link),
			chromedp.WaitReady("body"),
			chromedp.Sleep(10 * time.Second), // Wait for page load
		})
		if err != nil {
			log.Printf("Failed to navigate to auth URL: %v", err)
			return
		}
		log.Printf("Successfully navigated to auth URL")

		log.Printf("Handling cookie consent...")
		err = chromedp.Run(ctx, chromedp.Tasks{
			// Handle cookie consent with longer timeout
			chromedp.WaitVisible(`#onetrust-accept-btn-handler`, chromedp.ByID),
			chromedp.Sleep(1 * time.Second), // Wait for animation
			chromedp.Click(`#onetrust-accept-btn-handler`, chromedp.ByID),
			chromedp.Sleep(1 * time.Second), // Wait for cookie banner to disappear
		})
		if err != nil {
			log.Printf("Failed to handle cookie consent: %v", err)
			return
		}
		log.Printf("Cookie consent handled successfully")

		log.Printf("Starting enterprise sign-in flow...")
		err = chromedp.Run(ctx, chromedp.Tasks{
			// Enterprise sign-in flow
			chromedp.WaitVisible(`#enterprise-sign-in`, chromedp.ByID),
			chromedp.Click(`#enterprise-sign-in`, chromedp.ByID),
			chromedp.Sleep(1 * time.Second), // Wait for transition
		})
		if err != nil {
			log.Printf("Failed to start enterprise sign-in: %v", err)
			return
		}
		log.Printf("Enterprise sign-in flow started")

		log.Printf("Waiting for enterprise input page...")
		err = chromedp.Run(ctx, chromedp.Tasks{
			// Wait for page load to complete
			chromedp.WaitReady("body"),
			chromedp.Sleep(1 * time.Second), // Additional wait for dynamic content
			// Wait for enterprise input page and enter credentials
			chromedp.WaitVisible(`div[data-cyid="sign-in-with-enterprise"] input`),
			chromedp.WaitEnabled(`div[data-cyid="sign-in-with-enterprise"] input`),
			chromedp.SendKeys(`div[data-cyid="sign-in-with-enterprise"] input`, TEST_USER_NAME),
			chromedp.Sleep(500 * time.Millisecond), // Wait for button state update
			chromedp.WaitEnabled(`button[data-cyid="sign-in-with-enterprise-continue-button"]`),
			chromedp.Click(`button[data-cyid="sign-in-with-enterprise-continue-button"]`),
			chromedp.Sleep(2 * time.Second), // Wait for initial redirect
		})
		if err != nil {
			log.Printf("Failed to enter enterprise credentials: %v", err)
			return
		}
		log.Printf("Enterprise credentials entered successfully")

		log.Printf("Waiting for redirect to complete...")
		err = chromedp.Run(ctx, chromedp.Tasks{
			// Wait for first redirect to complete
			chromedp.WaitReady("body"),
			chromedp.Sleep(2 * time.Second),
		})
		if err != nil {
			log.Printf("Failed to wait for redirect: %v", err)
			return
		}
		log.Printf("Redirect completed successfully")

		log.Printf("Handling login form...")
		err = chromedp.Run(ctx, chromedp.Tasks{
			// Handle login form with longer timeouts and more robust selectors
			chromedp.WaitVisible(`[data-testid="login-page-username-input"]`),
			chromedp.SendKeys(`[data-testid="login-page-username-input"]`, TEST_USER_NAME),
			chromedp.Sleep(1 * time.Second),

			chromedp.WaitVisible(`[data-testid="login-page-password-input"]`),
			chromedp.SendKeys(`[data-testid="login-page-password-input"]`, TEST_USER_PASS),
			chromedp.Sleep(1 * time.Second),
		})
		if err != nil {
			log.Printf("Failed to fill login form: %v", err)
			return
		}
		log.Printf("Login form filled successfully")

		log.Printf("Submitting login form...")
		err = chromedp.Run(ctx, chromedp.Tasks{
			// Click sign in button and wait longer for reCAPTCHA processing
			chromedp.WaitVisible(`[data-testid="login-page-continue-login-button"]`),
			chromedp.Click(`[data-testid="login-page-continue-login-button"]`),
			chromedp.Sleep(5 * time.Second),

			// Wait for recaptcha and final redirect
			chromedp.Sleep(5 * time.Second),
			page.StopScreencast(),
		})
		if err != nil {
			log.Printf("Failed to submit login form: %v", err)
			return
		}
		log.Printf("Login form submitted successfully, waiting for completion...")
	}()

	log.Printf("Waiting for auth code callback on port %d...", port)
	usrInfo, isNewUser, err := auth.HandleAuthCode(port, nil)

	if err != nil {
		log.Printf("Failed to handle auth code: %v", err)
		t.Fatal(err)
	}

	log.Printf("Auth code handled successfully")
	log.Printf("User info received - Email: %s, IsNewUser: %v", usrInfo.UserEmail, isNewUser)

	log.Printf("Validating user email...")
	if ENV == "stage" {
		log.Printf("Stage environment - Expected email: %s, Got: %s", STAGE_USER_EMAIL, usrInfo.UserEmail)
		if usrInfo.UserEmail != STAGE_USER_EMAIL {
			t.Errorf("Expected: %s, Got: %s", STAGE_USER_EMAIL, usrInfo.UserEmail)
		}
	} else {
		log.Printf("Production environment - Expected email: %s, Got: %s", USER_EMAIL, usrInfo.UserEmail)
		if usrInfo.UserEmail != USER_EMAIL {
			t.Errorf("Expected: %s, Got: %s", USER_EMAIL, usrInfo.UserEmail)
		}
	}

	log.Printf("Checking login status...")
	isLoggedIn := auth.IsLoggedIn()
	log.Printf("Login status: %v", isLoggedIn)
	if isLoggedIn != true {
		t.Errorf("auth.isLoggedIn() is not %v, Got: %v", true, isLoggedIn)
	}

	log.Printf("=== Login Test Completed Successfully ===")
}
