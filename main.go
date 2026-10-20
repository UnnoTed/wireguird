//go:generate go run ./gui/get/generator/generator.go -target=./gui/get
//go:generate go run github.com/UnnoTed/fileb0x fileb0x.toml
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/UnnoTed/go-appindicator"
	"github.com/UnnoTed/horizontal"
	"github.com/UnnoTed/wireguird/gui"
	"github.com/UnnoTed/wireguird/gui/countries"
	"github.com/UnnoTed/wireguird/static"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
)

var iconTmpDir = os.TempDir()
var win *gtk.ApplicationWindow
var singleInstanceLock *os.File

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("\n\n\n\n")
		fmt.Println("===========================================================================")
		fmt.Println("===========================================================================")
		fmt.Println("===========================================================================")
		fmt.Println("===                                                                     ===")
		fmt.Println("===                                                                     ===")
		fmt.Println("=== Wireguird must be ran as root to access /etc/wireguard and wg-quick ===")
		fmt.Println("=== Use pkexec to run as root:                                          ===")
		fmt.Println("===  pkexec wireguird                                                   ===")
		fmt.Println("===                                                                     ===")
		fmt.Println("===                                                                     ===")
		fmt.Println("===========================================================================")
		fmt.Println("===========================================================================")
		fmt.Println("===========================================================================")
		return
	}

	defer os.Remove(iconTmpDir)
	defer countries.CloseDatabase()

	// gtk.Application's own single-instance handling relies on the session bus,
	// which is unavailable when running as root via pkexec, so a lock file is
	// used to stop instances from piling up.
	if !acquireSingleInstance() {
		return
	}

	log.Logger = log.Output(horizontal.ConsoleWriter{Out: os.Stderr})
	log.Info().Uint("major", gtk.GetMajorVersion()).Uint("minor", gtk.GetMinorVersion()).Uint("micro", gtk.GetMicroVersion()).Msg("GTK Version")

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

	// another launch (e.g. clicking the desktop icon again) asks the running
	// instance to bring its window back instead of starting a second process
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	go func() {
		for range sigCh {
			glib.IdleAdd(func() {
				if win != nil {
					win.Show()
					win.Present()
				}
			})
		}
	}()

	os.Exit(application.Run(os.Args))
}

// acquireSingleInstance takes an exclusive lock for the process lifetime and
// returns whether this process should run. If another instance holds the lock,
// it is asked to show its window and this process exits.
func acquireSingleInstance() bool {
	lockPath := "/run/wireguird.lock"

	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		lockPath = filepath.Join(os.TempDir(), "wireguird.lock")
		lock, err = os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			log.Error().Err(err).Msg("error opening single instance lock file")
			return true
		}
	}

	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if data, err := os.ReadFile(lockPath); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
				syscall.Kill(pid, syscall.SIGUSR1)
			}
		}

		fmt.Println("Wireguird is already running, showing its window.")
		return false
	}

	// record our pid so the next launch can ask us to show the window
	lock.Truncate(0)
	lock.WriteAt([]byte(strconv.Itoa(os.Getpid())+"\n"), 0)
	lock.Sync()

	// keep the file open so the lock is held until the process exits
	singleInstanceLock = lock

	return true
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

	iconsDir := "/usr/share/wireguird/Icon/"
	if _, err := os.Stat(iconsDir); os.IsNotExist(err) {
		iconsDir = iconTmpDir
		icons := []string{"wireguard_off", "wg_connected"}
		for _, icon := range icons {
			iconBytes, err := static.ReadFile("./Icon/" + icon + ".png")
			if err != nil {
				log.Error().Err(err).Str("icon", icon).Msg("Error reading embedded icon")
				return nil, err
			}

			iconPath := filepath.Join(iconTmpDir, icon+".png")
			if err := os.WriteFile(iconPath, iconBytes, 0644); err != nil {
				log.Error().Err(err).Str("icon", icon).Msg("Failed to write temp icon")
				return nil, err
			}
		}
	}

	indicator := appindicator.New(application.GetApplicationID(), "wireguard_off", appindicator.CategoryApplicationStatus)
	indicator.SetIconThemePath(iconsDir)
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

	// once the window has been laid out, grow it by the space the peer list
	// needs; the default size already fits a single peer
	fitted := false
	win.Connect("size-allocate", func() {
		if fitted {
			return
		}
		fitted = true
		fitWindowToPeers()
	})

	if !gui.Settings.StartOnTray {
		win.ShowAll()
	}

	win.SetTitle("Wireguird")

	return nil
}

// fitWindowToPeers grows the window so that all peers of the selected tunnel
// are visible without scrolling. If there are too many peers to fit, the
// window is made as tall as the monitor's work area and the peers scroll.
func fitWindowToPeers() {
	extra := gui.PeersExtraHeight()
	if extra <= 0 {
		return
	}

	width, height := win.GetSize()
	wanted := height + extra

	// don't grow beyond the monitor's work area
	maxHeight := 0
	if display, err := gdk.DisplayGetDefault(); err == nil && display != nil {
		monitor, merr := display.GetPrimaryMonitor()
		if (merr != nil || monitor == nil) && display.GetNMonitors() > 0 {
			monitor, _ = display.GetMonitor(0)
		}

		if monitor != nil {
			if area := monitor.GetWorkarea(); area != nil {
				_, _, _, h := area.GetRectangleInt()
				maxHeight = h
			}
		}
	}

	if maxHeight > 0 && wanted > maxHeight {
		wanted = maxHeight
	}

	win.Resize(width, wanted)
}
