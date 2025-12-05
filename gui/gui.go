package gui

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/UnnoTed/go-appindicator"
	"github.com/UnnoTed/wireguird/gui/get"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/wgctrl"
)

const (
	Version            = "1.1.0"
	Repo               = "https://github.com/UnnoTed/wireguird"
	DefaultTunnelsPath = "/etc/wireguard/"
	IconPath           = "/opt/wireguird/Icon/"
)

var (
	settingsWindow *gtk.Window
	editorWindow   *gtk.Window
	application    *gtk.Application
	indicator      *appindicator.Indicator
	builder        *gtk.Builder
	window         *gtk.ApplicationWindow
	header         *gtk.HeaderBar
	wgc            *wgctrl.Client
	updateTicker   *time.Ticker
	TunnelsPath    string
)

func Create(app *gtk.Application, b *gtk.Builder, w *gtk.ApplicationWindow, ind *appindicator.Indicator) error {
	application = app
	get.Builder = b
	indicator = ind
	builder = b
	window = w

	var err error
	header, err = get.HeaderBar("main_header")
	if err != nil {
		return err
	}

	wgc, err = wgctrl.New()
	if err != nil {
		ShowError(w, err)
		return err
	}

	ds, err := wgc.Devices()
	if err != nil {
		ShowError(w, err)
		return err
	}

	indicator.SetIcon("wireguard_off")

	for _, d := range ds {
		header.SetSubtitle("Connected to " + d.Name)
		indicator.SetIcon("wg_connected")
	}

	if _, err := createEditor("", false); err != nil {
		return err
	}

	if _, err := createSettings(false); err != nil {
		return err
	}

	t := &Tunnels{}
	if err := t.Create(); err != nil {
		return err
	}

	if Settings.CheckUpdates {
		go func() {
			time.Sleep(60 * time.Second)
			if err := updateCheck(); err != nil {
				log.Error().Err(err).Msg("error on update check")
			}

			updateTicker = time.NewTicker(24 * time.Hour)
			defer updateTicker.Stop()

			for {
				select {
				case <-updateTicker.C:
					if err := updateCheck(); err != nil {
						log.Error().Err(err).Msg("error on update check")
					}
				}
			}
		}()
	}

	window.HideOnDelete()
	return nil
}

func ShowError(win *gtk.ApplicationWindow, err error, info ...string) {
	if err == nil {
		return
	}

	glib.IdleAdd(func() {
		wlog("ERROR", err.Error())
		dlg := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", err.Error())
		dlg.SetTitle("Error")
		dlg.Run()
		dlg.Destroy()
	})
}

func createEditor(name string, show bool) (*gtk.Window, error) {
	if editorWindow == nil {
		var err error
		editorWindow, err = get.Window("editor_window")
		if err != nil {
			log.Error().Err(err).Msg("error getting main_window")
			return nil, err
		}

		// prevents having to re-create the editor window
		editorWindow.HideOnDelete()
	}

	if show {
		ne, err := get.Entry("editor_name")
		if err != nil {
			return nil, err
		}
		ne.SetText(name)

		editorWindow.SetTitle("Edit tunnel - " + name)
		editorWindow.ShowAll()
	}

	return editorWindow, nil
}

func createSettings(show bool) (*gtk.Window, error) {
	if settingsWindow == nil {
		var err error
		settingsWindow, err = get.Window("settings_window")
		if err != nil {
			log.Error().Err(err).Msg("error getting main_window")
			return nil, err
		}

		// prevents having to re-create the settings window
		settingsWindow.HideOnDelete()
	}

	if show {
		settingsWindow.SetTitle("Wireguird - Settings")
		settingsWindow.ShowAll()
	}

	return settingsWindow, nil
}

func updateCheck() error {
	log.Info().Msg("Checking for updates")
	resp, err := http.Get("https://api.github.com/repos/UnnoTed/wireguird/releases")
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	j := []map[string]interface{}{}
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	log.Info().Interface("json", j).Msg("response")

	if len(j) > 0 {
		latest := j[0]

		if tagName, ok := latest["tag_name"].(string); ok && tagName != "v"+Version {
			glib.IdleAdd(func() {
				d := gtk.MessageDialogNew(window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "Wireguird: update available => "+tagName)
				defer d.Destroy()

				d.Run()
			})
		}
	}

	return nil
}
