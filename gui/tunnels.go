package gui

import (
	"archive/zip"
	"fmt"
	"github.com/ungerik/go-dry"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UnnoTed/wireguird/gui/get"
	"github.com/dustin/go-humanize"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

var Connected = false

type Tunnels struct {
	Interface struct {
		Status     *gtk.Label
		PublicKey  *gtk.Label
		ListenPort *gtk.Label
		Addresses  *gtk.Label
		DNS        *gtk.Label
	}

	Peer struct {
		PublicKey       *gtk.Label
		AllowedIPs      *gtk.Label
		Endpoint        *gtk.Label
		LatestHandshake *gtk.Label
		Transfer        *gtk.Label
	}

	ButtonChangeState *gtk.Button
	icons             map[string]*gtk.Image
	ticker            *time.Ticker

	lastSelected string
}

func (t *Tunnels) Create() error {
	t.icons = map[string]*gtk.Image{}
	t.ticker = time.NewTicker(1 * time.Second)

	tl, err := get.ListBox("tunnel_list")
	if err != nil {
		return err
	}

	if err := t.ScanTunnels(); err != nil {
		return err
	}

	// menu
	{
		mb, err := get.MenuButton("menu")
		if err != nil {
			return err
		}

		menu, err := gtk.MenuNew()
		if err != nil {
			return err
		}

		mVersion, err := gtk.MenuItemNew()
		if err != nil {
			return err
		}

		mVersion.SetLabel("VERSION: v" + Version)
		//mVersion.SetSensitive(false)
		mVersion.Connect("activate", func() {
			if err := exec.Command("xdg-open", Repo).Start(); err != nil {
				ShowError(window, err, "open repo url error")
			}
		})
		menu.Append(mVersion)

		menu.SetHAlign(gtk.ALIGN_CENTER)
		menu.SetVAlign(gtk.ALIGN_CENTER)

		menu.ShowAll()
		mb.SetPopup(menu)
	}

	t.Interface.Status, err = get.Label("label_interface_status")
	if err != nil {
		return err
	}

	t.Interface.PublicKey, err = get.Label("label_interface_public_key")
	if err != nil {
		return err
	}

	t.Interface.ListenPort, err = get.Label("label_interface_listen_port")
	if err != nil {
		return err
	}

	t.Interface.Addresses, err = get.Label("label_interface_addresses")
	if err != nil {
		return err
	}

	t.Interface.DNS, err = get.Label("label_interface_dns_servers")
	if err != nil {
		return err
	}

	t.Peer.PublicKey, err = get.Label("label_peer_public_key")
	if err != nil {
		return err
	}

	t.Peer.AllowedIPs, err = get.Label("label_peer_allowed_ips")
	if err != nil {
		return err
	}

	t.Peer.Endpoint, err = get.Label("label_peer_endpoint")
	if err != nil {
		return err
	}

	t.Peer.LatestHandshake, err = get.Label("label_peer_latest_handshake")
	if err != nil {
		return err
	}

	t.Peer.Transfer, err = get.Label("label_peer_transfer")
	if err != nil {
		return err
	}

	t.ButtonChangeState, err = get.Button("button_change_state")
	if err != nil {
		return err
	}

	t.ButtonChangeState.Connect("clicked", func() {
		err := func() error {
			list, err := wgc.Devices()
			if err != nil {
				return err
			}

			activeName := t.ActiveDeviceName()
			for _, d := range list {
				gray, err := gtk.ImageNewFromFile("/opt/wireguird/Icon/not_connected.png")
				if err != nil {
					return err
				}

				glib.IdleAdd(func() {
					t.icons[d.Name].SetFromPixbuf(gray.GetPixbuf())
				})

				if err := exec.Command("wg-quick", "down", d.Name).Run(); err != nil {
					return err
				}

				indicator.SetIcon("wireguard_off")
			}

			row := tl.GetSelectedRow()
			// row not found for config
			if row == nil {
				return nil
			}

			name, err := row.GetName()
			if err != nil {
				return err
			}

			// dont connect to the new one
			if activeName != "" && activeName == name {
				t.UpdateRow(row)

				glib.IdleAdd(func() {
					header.SetSubtitle("Not connected!")
				})
				return nil
			}

			if err := exec.Command("wg-quick", "up", name).Run(); err != nil {
				return err
			}

			glib.IdleAdd(func() {
				header.SetSubtitle("Connected to " + name)
			})

			green, err := gtk.ImageNewFromFile("/opt/wireguird/Icon/connected.png")
			if err != nil {
				return err
			}

			glib.IdleAdd(func() {
				t.icons[name].SetFromPixbuf(green.GetPixbuf())
				t.UpdateRow(row)
				indicator.SetIcon("wg_connected")
			})

			if err := wlog("INFO", "Connected to "+name); err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			ShowError(window, err)
		}
	})

	// boxPeers, err := get.Box("box_peers")
	// if err != nil {
	// 	return err
	// }

	tl.Connect("row-activated", func(l *gtk.ListBox, row *gtk.ListBoxRow) {
		t.UpdateRow(row)
	})

	// button: add tunnel
	btnAddTunnel, err := get.Button("button_add_tunnel")
	if err != nil {
		return err
	}

	btnAddTunnel.Connect("clicked", func() {
		err := func() error {
			log.Print("btn add tunnel")
			dialog, err := gtk.FileChooserNativeDialogNew("Wireguird - Choose a tunnel file (*.conf)", window, gtk.FILE_CHOOSER_ACTION_OPEN, "OK", "Cancel")
			if err != nil {
				return err
			}
			defer dialog.Destroy()

			// filter *.conf files
			filter, err := gtk.FileFilterNew()
			if err != nil {
				return err
			}
			filter.AddPattern("*.conf")
			filter.SetName("*.conf")
			dialog.AddFilter(filter)

			res := dialog.Run()
			if gtk.ResponseType(res) == gtk.RESPONSE_ACCEPT {
				fname := dialog.GetFilename()
				log.Print(fname)

				data, err := ioutil.ReadFile(fname)
				if err != nil {
					return err
				}

				err = ioutil.WriteFile(filepath.Join(TunnelsPath, filepath.Base(fname)), data, 666)
				if err != nil {
					return err
				}

				if err := t.ScanTunnels(); err != nil {
					return err
				}
			}

			return nil
		}()

		if err != nil {
			ShowError(window, err, "add tunnel error")
		}
	})

	// button: delete tunnel
	btnDelTunnel, err := get.Button("button_del_tunnel")
	if err != nil {
		return err
	}

	btnDelTunnel.Connect("clicked", func() {
		err := func() error {
			row := tl.GetSelectedRow()
			name, err := row.GetName()
			if err != nil {
				return err
			}

			ext := ""
			if !strings.HasSuffix(name, ".conf") {
				ext = ".conf"
			}

			path := filepath.Join(TunnelsPath, name+ext)

			d := gtk.MessageDialogNew(window, gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "Do you really want to delete "+name+"?")
			defer d.Destroy()

			res := d.Run()
			if res == gtk.RESPONSE_YES {
				if err := os.Remove(path); err != nil {
					return err
				}

				if err := t.ScanTunnels(); err != nil {
					return err
				}

				return nil
			}

			return nil
		}()

		if err != nil {
			ShowError(window, err, "tunnel delete error")
		}
	})

	// button: zip tunnels
	btnZipTunnels, err := get.Button("button_zip_tunnel")
	if err != nil {
		return err
	}

	btnZipTunnels.Connect("clicked", func() {
		err := func() error {
			d, err := gtk.FileChooserDialogNewWith2Buttons("Wireguird - zip tunnels -> Save zip file", window, gtk.FILE_CHOOSER_ACTION_SAVE, "Cancel", gtk.RESPONSE_CANCEL, "Save", gtk.RESPONSE_ACCEPT)
			if err != nil {
				panic(err)
			}
			defer d.Destroy()

			d.SetDoOverwriteConfirmation(true)
			t := time.Now()
			d.SetCurrentName(fmt.Sprint("wg_tunnels_", t.Day(), "_", t.Month(), "_", t.Year(), ".zip"))

			res := d.Run()
			if res == gtk.RESPONSE_ACCEPT {
				dest := d.GetFilename()
				base := strings.TrimSuffix(filepath.Base(dest), filepath.Ext(dest))

				zf, err := os.Create(dest)
				if err != nil {
					return err
				}
				defer zf.Close()

				archive := zip.NewWriter(zf)
				defer archive.Close()

				return filepath.Walk(TunnelsPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					header, err := zip.FileInfoHeader(info)
					if err != nil {
						return err
					}

					header.Name = filepath.Join(base, strings.TrimPrefix(path, TunnelsPath))

					if info.IsDir() {
						header.Name += "/"
					} else {
						header.Method = zip.Deflate
						header.SetMode(0777)
					}

					log.Debug().Interface("header", header).Interface("info", info).Msg("walk")
					writer, err := archive.CreateHeader(header)
					if err != nil {
						return err
					}

					if info.IsDir() {
						return nil
					}

					f, err := os.Open(path)
					if err != nil {
						return err
					}

					defer f.Close()
					if _, err = io.Copy(writer, f); err != nil {
						return err
					}

					return nil
				})
			}

			return nil
		}()

		if err != nil {
			ShowError(window, err, "tunnel delete error")
		}
	})

	// stores a modified state for the editor
	modified := false
	editorWindow.Connect("hide", func() {
		if err := t.ScanTunnels(); err != nil {
			ShowError(window, err, "scan tunnel after closing editor window error")
		}

		modified = false
	})

	ne, err := get.Entry("editor_name")
	if err != nil {
		return err
	}
	ne.Connect("changed", func() {
		modified = true
	})

	et, err := get.TextView("editor_text")
	if err != nil {
		return err
	}

	etb, err := et.GetBuffer()
	if err != nil {
		return err
	}

	etb.Connect("changed", func() {
		modified = true
	})

	// button: edit tunnel
	btnEditTunnel, err := get.Button("button_edit_tunnel")
	if err != nil {
		return err
	}

	btnEditTunnel.Connect("clicked", func() {
		modified = false

		err := func() error {
			row := tl.GetSelectedRow()
			if row == nil {
				return nil
			}

			name, err := row.GetName()
			if err != nil {
				return err
			}

			ext := ""
			if !strings.HasSuffix(name, ".conf") {
				ext = ".conf"
			}

			path := filepath.Join(TunnelsPath, name+ext)
			log.Print(path)

			ew, err := createEditor(name, true)
			if err != nil {
				return err
			}

			et, err := get.TextView("editor_text")
			if err != nil {
				return err
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// low budget GtkSourceView
			propertyColor := "purple"
			sectionColor := "green"

			conf := string(data)

			// gets public key to use in the PublicKey entry
			re := regexp.MustCompile(`(?m)PublicKey[\s]+=[\s]+(.*)`)
			m := re.FindStringSubmatch(conf)
			if len(m) >= 2 {
				epk, err := get.Entry("editor_publickey")
				if err != nil {
					return err
				}

				epk.SetText(m[1])
			}

			r := strings.NewReplacer(
				"&", "&amp;",
				"PrivateKey", "<span color=\""+propertyColor+"\">PrivateKey</span>",
				"PublicKey", "<span color=\""+propertyColor+"\">PublicKey</span>",
				"Address", "<span color=\""+propertyColor+"\">Address</span>",
				"DNS", "<span color=\""+propertyColor+"\">DNS</span>",
				"AllowedIPs", "<span color=\""+propertyColor+"\">AllowedIPs</span>",
				"Endpoint", "<span color=\""+propertyColor+"\">Endpoint</span>",
				"PostUp", "<span color=\""+propertyColor+"\">PostUp</span>",
				"PreDown", "<span color=\""+propertyColor+"\">PreDown</span>",
				"PreSharedKey", "<span color=\""+propertyColor+"\">PreSharedKey</span>",
				"PersistentKeepalive", "<span color=\""+propertyColor+"\">PersistentKeepalive</span>",
				"[Interface]", "<span color=\""+sectionColor+"\">[Interface]</span>",
				"[Peer]", "<span color=\""+sectionColor+"\">[Peer]</span>",
			)
			conf = r.Replace(conf)

			ttt, err := gtk.TextTagTableNew()
			if err != nil {
				return err
			}
			tb, err := gtk.TextBufferNew(ttt)
			if err != nil {
				return err
			}

			tb.InsertMarkup(tb.GetStartIter(), conf)
			et.SetBuffer(tb)

			ew.ShowAll()

			return nil
		}()

		if err != nil {
			ShowError(window, err, "edit tunnel error")
		}
	})

	// button: editor's cancel
	btnEditorCancel, err := get.Button("editor_button_cancel")
	if err != nil {
		return err
	}

	btnEditorCancel.Connect("clicked", func() {
		if !modified {
			editorWindow.Close()
			return
		}

		err := func() error {
			d := gtk.MessageDialogNew(editorWindow, gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "Do you want to cancel any modification done to the tunnel?")
			defer d.Destroy()

			res := d.Run()
			if res == gtk.RESPONSE_YES {
				modified = false
				editorWindow.Close()
				return nil
			}

			return nil
		}()

		if err != nil {
			ShowError(window, err, "cancel tunnel error")
		}
	})

	// button: editor's save
	btnEditorSave, err := get.Button("editor_button_save")
	if err != nil {
		return err
	}

	btnEditorSave.Connect("clicked", func() {
		err := func() error {
			row := tl.GetSelectedRow()
			if row == nil {
				return nil
			}

			name, err := row.GetName()
			if err != nil {
				return err
			}

			// adds extension when empty
			ext := ""
			if !strings.HasSuffix(name, ".conf") {
				ext = ".conf"
			}

			path := filepath.Join(TunnelsPath, name+ext)

			// get the new tunnel name
			nameEntry, err := get.Entry("editor_name")
			if err != nil {
				return err
			}

			newName, err := nameEntry.GetText()
			if err != nil {
				return err
			}
			newName = strings.TrimSpace(newName)

			// rename the tunnel
			if name != newName {
				newPath := filepath.Join(TunnelsPath, newName+ext)

				if err := os.Rename(path, newPath); err != nil {
					return err
				}

				path = newPath
			}

			// get the tunnels' edited text through editor_text's buffer
			etxt, err := get.TextView("editor_text")
			if err != nil {
				return err
			}

			b, err := etxt.GetBuffer()
			if err != nil {
				return err
			}

			data, err := b.GetText(b.GetStartIter(), b.GetEndIter(), false)
			if err != nil {
				return err
			}

			// write changes
			if err := ioutil.WriteFile(path, []byte(data), 666); err != nil {
				return err
			}

			modified = false
			if err := t.ScanTunnels(); err != nil {
				return err
			}

			editorWindow.Close()

			return nil
		}()

		if err != nil {
			ShowError(window, err, "save tunnel error")
		}
	})

	go func() {
		for {
			<-t.ticker.C

			if !window.HasToplevelFocus() {
				continue
			}

			row := tl.GetSelectedRow()
			if row == nil {
				t.UnknownLabels()
				continue
			}

			name, err := row.GetName()
			if err != nil {
				log.Error().Err(err).Msg("row get name err")
				continue
			}

			if name != t.ActiveDeviceName() {
				continue
			}

			d, err := wgc.Device(name)
			if err != nil {
				log.Error().Err(err).Msg("wgc get device err")
				continue
			}

			t.Interface.PublicKey.SetText(d.PublicKey.String())
			t.Interface.ListenPort.SetText(strconv.Itoa(d.ListenPort))

			for _, p := range d.Peers {
				hs := humanize.Time(p.LastHandshakeTime)
				glib.IdleAdd(func() {
					t.Peer.LatestHandshake.SetText(hs)
					t.Peer.Transfer.SetText(humanize.Bytes(uint64(p.ReceiveBytes)) + " received, " + humanize.Bytes(uint64(p.TransmitBytes)) + " sent")
				})
			}
		}
	}()

	return nil
}
func (t *Tunnels) UpdateRow(row *gtk.ListBoxRow) {
	err := func() error {
		ds, err := wgc.Devices()
		if err != nil {
			return err
		}

		log.Debug().Interface("list", ds).Msg("devices")

		id, err := row.GetName()
		if err != nil {
			return err
		}

		cfg, err := ini.Load(TunnelsPath + id + ".conf")
		if err != nil {
			return err
		}

		t.lastSelected = id
		t.UnknownLabels()

		peersec := cfg.Section("Peer")
		insec := cfg.Section("Interface")

		glib.IdleAdd(func() {
			t.Interface.Addresses.SetText(insec.Key("Address").String())
			t.Interface.Status.SetText("Inactive")
			t.Interface.DNS.SetText(insec.Key("DNS").String())

			t.ButtonChangeState.SetLabel("Activate")

			t.Peer.AllowedIPs.SetText(peersec.Key("AllowedIPs").String())
			t.Peer.PublicKey.SetText(peersec.Key("PublicKey").String())
			t.Peer.Endpoint.SetText(peersec.Key("Endpoint").String())
		})

		for _, d := range ds {
			if d.Name != id {
				continue
			}

			// i'll do this later
			// _ = boxPeers
			// boxPeers.GetChildren().Foreach(func(item interface{}) {
			// 	boxPeers.Remove(item.(*gtk.Widget))
			// })

			glib.IdleAdd(func() {
				t.Interface.Status.SetText("Active")
				t.ButtonChangeState.SetLabel("Deactivate")
				t.Interface.PublicKey.SetText(d.PublicKey.String())
				t.Interface.ListenPort.SetText(strconv.Itoa(d.ListenPort))
			})

			for _, p := range d.Peers {
				hs := humanize.Time(p.LastHandshakeTime)

				glib.IdleAdd(func() {
					t.Peer.LatestHandshake.SetText(hs)
					t.Peer.Transfer.SetText(humanize.Bytes(uint64(p.ReceiveBytes)) + " received, " + humanize.Bytes(uint64(p.TransmitBytes)) + " sent")
				})
			}

			break
		}

		return nil
	}()

	if err != nil {
		log.Error().Err(err).Msg("row activated")
		ShowError(window, err)
	}
}

func (t *Tunnels) ScanTunnels() error {
	var err error
	var configList []string
	list, err := dry.ListDirFiles(TunnelsPath)
	if err != nil {
		// showError(err)
		return err
	}

	for _, fileName := range list {
		if !strings.HasSuffix(fileName, ".conf") {
			continue
		}

		configList = append(configList, strings.TrimSuffix(fileName, ".conf"))
	}

	tl, err := get.ListBox("tunnel_list")
	if err != nil {
		return err
	}

	for {
		row := tl.GetRowAtIndex(0)
		if row == nil {
			break
		}

		row.Destroy()
	}

	sort.Slice(configList, func(a, b int) bool {
		return configList[a] < configList[b]
	})

	activeName := t.ActiveDeviceName()

	lasti := len(configList) - 1
	for i, name := range configList {
		row, err := gtk.ListBoxRowNew()
		if err != nil {
			return err
		}
		row.SetName(name)
		row.SetMarginStart(8)
		row.SetMarginEnd(8)
		if i == 0 {
			row.SetMarginTop(8)
		} else if i == lasti {
			row.SetMarginBottom(8)
		}

		// icon, err := gtk.ButtonNew()
		// icon, err := gtk.ColorButtonNew()
		// if err != nil {
		// 	return err
		// }

		var img *gtk.Image

		if activeName == name {
			green, err := gtk.ImageNewFromFile("/opt/wireguird/Icon/connected.png")
			if err != nil {
				return err
			}

			img = green
		} else {
			gray, err := gtk.ImageNewFromFile("/opt/wireguird/Icon/not_connected.png")
			if err != nil {
				return err
			}
			img = gray
		}

		t.icons[name] = img
		// img, err := gtk.ImageNewFromFile("/opt/wireguird/Icon/not_connected.png")
		// if err != nil {
		// 	return err
		// }

		// icon.SetImage(img)

		// dg, err := static.ReadFile("icon/dot-gray.svg")
		// if err != nil {
		// 	return err
		// }

		// pbl, err := gdk.PixbufLoaderNew()
		// if err != nil {
		// 	return err
		// }

		// pb, err := pbl.WriteAndReturnPixbuf(dg)
		// if err != nil {
		// 	return err
		// }

		// img, err := gtk.ImageNewFromPixbuf(pb)
		// if err != nil {
		// 	return err
		// }

		// icon.Image
		// icon.SetImage(img)

		img.SetVAlign(gtk.ALIGN_CENTER)
		img.SetHAlign(gtk.ALIGN_START)
		img.SetSizeRequest(10, 10)
		// icon.SetMarginBottom(0)
		// icon.SetMarginTop(0)
		img.SetVExpand(false)
		img.SetHExpand(false)

		// sctx, err := icon.GetStyleContext()
		// if err != nil {
		// 	return err
		// }
		// sctx.AddClass("circular")

		// if name == activeName {
		// 	sctx.AddClass("btn-green")
		// 	// rgba := gdk.NewRGBA(102, 204, 153, 1)

		// }

		label, err := gtk.LabelNew(name)
		if err != nil {
			return err
		}
		label.SetHAlign(gtk.ALIGN_START)
		// label.SetHExpand(true)

		label.SetMarginStart(8)
		label.SetMarginEnd(8)
		label.SetMarginTop(8)
		label.SetMarginBottom(8)

		box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 4)
		if err != nil {
			return err
		}

		box.Add(img)
		box.Add(label)

		row.SetName(name)
		row.Add(box)
		row.ShowAll()
		tl.Insert(row, -1)

		if name == t.lastSelected {
			tl.SelectRow(row)
			t.UpdateRow(row)
		}
	}

	return nil
}

func (t *Tunnels) UnknownLabels() {
	glib.IdleAdd(func() {
		t.ButtonChangeState.SetLabel("unknown")

		t.Interface.Addresses.SetText("unknown")
		t.Interface.Status.SetText("unknown")
		t.Interface.DNS.SetText("unknown")

		t.Peer.AllowedIPs.SetText("unknown")
		t.Peer.PublicKey.SetText("unknown")
		t.Peer.Endpoint.SetText("unknown")

		t.Interface.Status.SetText("unknown")
		t.ButtonChangeState.SetLabel("unknown")
		t.Interface.PublicKey.SetText("unknown")
		t.Interface.ListenPort.SetText("unknown")
		t.Peer.LatestHandshake.SetText("unknown")

		t.Peer.Transfer.SetText("unknown")
	})
}

func (t *Tunnels) ActiveDeviceName() string {
	ds, _ := wgc.Devices()

	for _, d := range ds {
		return d.Name
	}

	return ""
}

func wlog(t string, text string) error {
	wlogs, err := get.ListBox("wireguard_logs")
	if err != nil {
		return err
	}

	l, err := gtk.LabelNew("")
	if err != nil {
		return err
	}

	if t == "ERROR" {
		t = `<span color="#FF0000">` + t + "</span>"
	}

	l.SetMarkup(`<span color="#008080">[` + time.Now().Format("02/Jan/06 15:04:05 MST") + `]</span>[` + t + `]: ` + text)
	l.SetHExpand(true)
	l.SetHAlign(gtk.ALIGN_START)

	glib.IdleAdd(func() {
		l.Show()
		wlogs.Add(l)
	})

	return nil
}
