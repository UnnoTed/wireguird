package gui

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UnnoTed/wireguird/gui/get"
	"github.com/dustin/go-humanize"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
	"github.com/ungerik/go-dry"
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
}

func (t *Tunnels) Create() error {
	t.icons = map[string]*gtk.Image{}
	t.ticker = time.NewTicker(1 * time.Second)

	var configList []string
	list, err := dry.ListDirFiles("/etc/wireguard/")
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
			green, err := gtk.ImageNewFromFile("./icon/dot-green.png")
			if err != nil {
				return err
			}

			img = green
		} else {
			gray, err := gtk.ImageNewFromFile("./icon/dot-gray.png")
			if err != nil {
				return err
			}
			img = gray
		}

		t.icons[name] = img
		// img, err := gtk.ImageNewFromFile("./icon/dot-gray.png")
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

			activeName = t.ActiveDeviceName()
			for _, d := range list {
				gray, err := gtk.ImageNewFromFile("./icon/dot-gray.png")
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
			name, err := row.GetName()
			if err != nil {
				return err
			}

			// dont connect to the new one
			if activeName != "" && activeName == name {
				t.UpdateRow(row)
				header.SetSubtitle("Not connected!")
				return nil
			}

			if err := exec.Command("wg-quick", "up", name).Run(); err != nil {
				return err
			}

			glib.IdleAdd(func() {
				header.SetSubtitle("Connected to " + name)
			})

			green, err := gtk.ImageNewFromFile("./icon/dot-green.png")
			if err != nil {
				return err
			}

			glib.IdleAdd(func() {
				t.icons[name].SetFromPixbuf(green.GetPixbuf())
				t.UpdateRow(row)
				indicator.SetIcon("wireguard")
			})
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

	go func() {
		for {
			<-t.ticker.C

			if !window.HasToplevelFocus() {
				continue
			}

			row := tl.GetSelectedRow()
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
				t.Peer.LatestHandshake.SetText(hs)

				t.Peer.Transfer.SetText(humanize.Bytes(uint64(p.ReceiveBytes)) + " received, " + humanize.Bytes(uint64(p.TransmitBytes)) + " sent")
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

		cfg, err := ini.Load("/etc/wireguard/" + id + ".conf")
		if err != nil {
			return err
		}

		t.UnknownLabels()

		peersec := cfg.Section("Peer")
		insec := cfg.Section("Interface")

		t.Interface.Addresses.SetText(insec.Key("Address").String())
		t.Interface.Status.SetText("Inactive")
		t.Interface.DNS.SetText(insec.Key("DNS").String())

		t.ButtonChangeState.SetLabel("Activate")

		t.Peer.AllowedIPs.SetText(peersec.Key("AllowedIPs").String())
		t.Peer.PublicKey.SetText(peersec.Key("PublicKey").String())
		t.Peer.Endpoint.SetText(peersec.Key("Endpoint").String())

		for _, d := range ds {
			if d.Name != id {
				continue
			}

			// i'll do this later
			// _ = boxPeers
			// boxPeers.GetChildren().Foreach(func(item interface{}) {
			// 	boxPeers.Remove(item.(*gtk.Widget))
			// })

			t.Interface.Status.SetText("Active")
			t.ButtonChangeState.SetLabel("Deactivate")
			t.Interface.PublicKey.SetText(d.PublicKey.String())
			t.Interface.ListenPort.SetText(strconv.Itoa(d.ListenPort))

			for _, p := range d.Peers {
				hs := humanize.Time(p.LastHandshakeTime)
				t.Peer.LatestHandshake.SetText(hs)

				t.Peer.Transfer.SetText(humanize.Bytes(uint64(p.ReceiveBytes)) + " received, " + humanize.Bytes(uint64(p.TransmitBytes)) + " sent")
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

func (t *Tunnels) UnknownLabels() {
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
}

func (t *Tunnels) ActiveDeviceName() string {
	ds, _ := wgc.Devices()

	for _, d := range ds {
		return d.Name
	}

	return ""
}
