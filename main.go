//go:generate go run ./gui/get/generator/generator.go -target=./gui/get
//go:generate go run github.com/UnnoTed/fileb0x fileb0x.toml
package main

import (
	"os"
	"strings"

	"github.com/UnnoTed/go-appindicator"
	"github.com/UnnoTed/horizontal"
	"github.com/UnnoTed/wireguird/gui"
	"github.com/UnnoTed/wireguird/static"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
)

var win *gtk.ApplicationWindow

func main() {
	log.Logger = log.Output(horizontal.ConsoleWriter{Out: os.Stderr})
	log.Info().Uint("major", gtk.GetMajorVersion()).Uint("minor", gtk.GetMinorVersion()).Uint("micro", gtk.GetMicroVersion()).Msg("GTK Version")

	if gui.Settings.TunnelsPath == "" {
		gui.TunnelsPath = gui.DefaultTunnelsPath
	} else {
		gui.TunnelsPath = gui.Settings.TunnelsPath
	}
	if !strings.HasSuffix(gui.TunnelsPath, "/") {
		gui.TunnelsPath += "/"
	}

	if err := gui.Settings.Load(); err != nil {
		log.Error().Err(err).Msg("error initial settings load")
	}

	const appID = "com.wireguard.desktop"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Error().Err(err).Msg("error creating application")
		return
	}

	application.Connect("activate", func() {
		if err := createWindow(application); err != nil {
			log.Error().Err(err).Msg("create window error")
		}
	})

	os.Exit(application.Run(os.Args))
}

func createTray(application *gtk.Application) (*appindicator.Indicator, error) {
	menu, err := gtk.MenuNew()
	if err != nil {
		return nil, err
	}

	menuShow, err := gtk.MenuItemNewWithLabel("Show")
	if err != nil {
		return nil, err
	}

	menuQuit, err := gtk.MenuItemNewWithLabel("Quit")
	if err != nil {
		return nil, err
	}

	indicator := appindicator.New(application.GetApplicationID(), "wireguard_off", appindicator.CategoryApplicationStatus)
	indicator.SetIconThemePath("/opt/wireguird/Icon")
	indicator.SetTitle("Wireguird")
	// indicator.SetLabel("Wireguird", "")
	indicator.SetStatus(appindicator.StatusActive)
	indicator.SetMenu(menu)

	menuShow.Connect("activate", func() {
		win.Show()
		// createWindow(application)
	})
	if err != nil {
		return nil, err
	}

	menuQuit.Connect("activate", func() {
		application.Quit()
	})

	menu.Add(menuShow)
	menu.Add(menuQuit)
	menu.ShowAll()

	return indicator, nil
}

func createWindow(application *gtk.Application) error {
	data, err := static.ReadFile("wireguird.glade")
	if err != nil {
		log.Error().Err(err).Msg("cant read wireguird.glade")
		return err
	}

	b, err := gtk.BuilderNew()
	b.AddFromString(string(data))
	if err != nil {
		log.Error().Err(err).Msg("error getting main glade file")
		return err
	}

	wobj, err := b.GetObject("main_window")
	if err != nil {
		log.Error().Err(err).Msg("error getting main_window")
		return err
	}

	var ok bool
	win, ok = wobj.(*gtk.ApplicationWindow)
	if !ok {
		panic("not window")
	}
	application.AddWindow(win)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// win.SetDecorated(false)

	css, err := gtk.CssProviderNew()
	if err != nil {
		log.Error().Err(err).Msg("error creating css provider")
		return err
	}

	css.LoadFromData(`

		`)
	// css.LoadFromPath("./style.css")
	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		log.Error().Err(err).Msg("error getting screen")
		return err
	}

	gtk.AddProviderForScreen(screen, css, 1)

	indicator, err := createTray(application)
	if err != nil {
		log.Error().Err(err).Msg("create tray error")
	}

	if err := gui.Create(application, b, win, indicator); err != nil {
		log.Error().Err(err).Msg("error gui setup")
		return err
	}

	if !gui.Settings.StartOnTray {
		win.ShowAll()
	}

	win.SetTitle("Wireguird")

	return nil
}
