// Browser open tool

package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func openBrowser(port int, ssl bool) {
	// Generate localhost URL
	var url string

	if ssl {
		if port == 443 {
			url = "https://localhost"
		} else {
			url = "https://localhost:" + fmt.Sprint(port)
		}
	} else {
		if port == 80 {
			url = "http://localhost"
		} else {
			url = "http://localhost:" + fmt.Sprint(port)
		}
	}

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LaunchingBrowser",
			Other: "Launching browser: {{.URL}}",
		},
		TemplateData: map[string]interface{}{
			"URL": url,
		},
	})
	fmt.Println(msg)

	// Open the browser
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorBrowser",
				Other: "Error: could not open browser: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
	}
}
