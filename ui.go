package main

import (
	"NordVPNGUI/nordvpn"
	"encoding/json"
	"fmt"
	"github.com/energye/energy/v2/cef"
	"github.com/energye/energy/v2/cef/ipc"
	"github.com/energye/energy/v2/consts"
	"github.com/energye/golcl/lcl"
	"io"
	"net/http"
	"os"
	"strings"
)

func open(debug bool) {
	cef.GlobalInit(nil, nil)
	app := cef.NewApplication()
	if debug {
		app.AddCustomCommandLine("remote-debugging-port", "8080")
	}
	app.AddCustomCommandLine("disable-web-security", "true")
	app.AddCustomCommandLine("mute-audio", "true")
	cwd, _ := os.Getwd()

	cef.BrowserWindow.Config.Title = "NordVPNGUI Linux"
	cef.BrowserWindow.Config.Icon = cwd + "/ui/icon/icon.png"
	cef.BrowserWindow.Config.Url = "file://" + cwd + "/ui/index.html"
	cef.BrowserWindow.Config.MinWidth = 860
	cef.BrowserWindow.Config.MinHeight = 600
	cef.BrowserWindow.Config.Width = 860
	cef.BrowserWindow.Config.Height = 600

	cef.BrowserWindow.SetBrowserInit(func(event *cef.BrowserEvent, window cef.IBrowserWindow) {
		login_in_progress := false
		var popup_window cef.IBrowserWindow = nil

		window.Chromium().SetOnBeforePopup(func(sender lcl.IObject, browser *cef.ICefBrowser, frame *cef.ICefFrame, beforePopupInfo *cef.BeforePopupInfo, popupFeatures *cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, resultClient *cef.ICefClient, settings *cef.TCefBrowserSettings, resultExtraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool) bool {

			if !login_in_progress {
				config := &cef.TCefChromiumConfig{}
				config.SetEnableWindowPopup(false)

				windowProp := cef.WindowProperty{
					Width:  800,
					Height: 600,
					Title:  "NordVPN Login",
					Icon:   cwd + "/ui/icon/icon.png",
					Url:    beforePopupInfo.TargetUrl,
				}

				newWindow := cef.NewBrowserWindow(config, windowProp, nil)

				newWindow.Chromium().SetOnClose(func(sender lcl.IObject, browser *cef.ICefBrowser, aAction *consts.TCefCloseBrowserAction) {
					login_in_progress = false
				})

				newWindow.Show()

				popup_window = newWindow
				login_in_progress = true
			}

			return true
		})

		event.SetOnBeforeResourceLoad(func(sender lcl.IObject, browser *cef.ICefBrowser, frame *cef.ICefFrame, request *cef.ICefRequest, callback *cef.ICefCallback, result *consts.TCefReturnValue, window cef.IBrowserWindow) {
			//https://nordaccount.com/product/nordvpn/login/success?return=1&redirect_upon_open=1&exchange_token=OTM3MmI0YzA3YzM5Y2NlNmI2Y2JhZjA4NWNjZjBlOWZkOGNiOGFiYjk1YjliZDY5Mjk4Yjk3NzNkYWFiMzVhMw%3D%3D

			if strings.Contains(request.URL(), "https://nordaccount.com/product/nordvpn/login/success") {
				if login_in_progress {
					popup_window.Close()
					popup_window = nil

					token := strings.Split(request.URL(), "exchange_token=")[1]

					println("Token: " + token)
				}
			}
		})
	})

	register_ipcs()

	cef.Run(app)
}

func register_ipcs() {
	activeSpecialty := ""
	activeCountry := ""

	ipc.On("get_account", func() string {
		call := nordvpn.PrepareCall()
		acc, err := call.GetAccount()
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		call.EndCall()
		out, err := json.Marshal(acc)
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		return string(out)
	})

	ipc.On("get_login_url", func() string {
		call := nordvpn.PrepareCall()
		url, err := call.GetLoginURL()
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		call.EndCall()
		return fmt.Sprintf("{\"Login\": { \"url\": \"%s\" }}", url)
	})

	ipc.On("get_countries", func() string {
		resp, err := http.Get("https://api.nordvpn.com/v1/servers?limit=200000")
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}

		return string(body)
	})

	ipc.On("connect", func(country string, city string, group string) string {
		if group == "obfuscated" {
			// User tries to connect to obfuscated server
			// nordvpn settigns set obfuscated on
			// nordvpn connect

			return ""
		}

		call := nordvpn.PrepareCall()
		resp, err := call.Connect(country, city, group)

		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}

		out, jsonErr := json.Marshal(strings.Split(resp, "|"))

		if jsonErr != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", jsonErr.Error())
		}

		activeSpecialty = group
		activeCountry = country

		return "{\"ConnectionInfo\": " + string(out) + "}"
	})

	ipc.On("disconnect", func() string {
		call := nordvpn.PrepareCall()
		resp, err := call.Disconnect()
		call.EndCall()
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		return resp
	})

	ipc.On("reconnect", func() string {
		call := nordvpn.PrepareCall()
		resp, err := call.Disconnect()
		call.EndCall()
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		if resp == "false" {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", "Unknown error")
		}
		call = nordvpn.PrepareCall()
		resp, err = call.Connect(activeCountry, "", activeSpecialty)
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		out, jsonErr := json.Marshal(strings.Split(resp, "|"))
		if jsonErr != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", jsonErr.Error())
		}
		return "{\"ConnectionInfo\": " + string(out) + "}"
	})

	ipc.On("get_status", func() string {
		call := nordvpn.PrepareCall()
		resp, err := call.GetStatus()
		call.EndCall()
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		out, jsonErr := json.Marshal(resp)
		if jsonErr != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", jsonErr.Error())
		}
		return string(out)
	})

	ipc.On("get_register_url", func() string {
		call := nordvpn.PrepareCall()
		url, err := call.Register()
		if err != nil {
			return fmt.Sprintf("{\"Error\": { \"error\": \"%s\" }}", err.Error())
		}
		call.EndCall()
		return fmt.Sprintf("{\"Login\": { \"url\": \"%s\" }}", url)
	})
}
