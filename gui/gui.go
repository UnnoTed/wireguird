package gui

import (
	"github.com/UnnoTed/wireguird/gui/get"
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/wgctrl"
)

var (
	application *gtk.Application
	indicator   *appindicator.Indicator
	builder     *gtk.Builder
	window      *gtk.ApplicationWindow
	header      *gtk.HeaderBar
	wgc         *wgctrl.Client
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
		indicator.SetIcon("wireguard")
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

	if _, err := glib.IdleAdd(func() {
		dlg := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", err.Error())
		dlg.SetTitle("Error")
		dlg.Run()
		dlg.Destroy()
	}); err != nil {
		log.Error().Err(err).Msg("cant run idleadd")
	}
}
