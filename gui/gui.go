package gui

import (
	"github.com/UnnoTed/wireguird/gui/get"
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/wgctrl"
)

const (
	Version     = "0.2.0"
	Repo        = "https://github.com/UnnoTed/wireguird"
	TunnelsPath = "/etc/wireguard/"
)

var (
	editorWindow *gtk.Window
	application  *gtk.Application
	indicator    *appindicator.Indicator
	builder      *gtk.Builder
	window       *gtk.ApplicationWindow
	header       *gtk.HeaderBar
	wgc          *wgctrl.Client
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

	t := &Tunnels{}
	if err := t.Create(); err != nil {
		return err
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
