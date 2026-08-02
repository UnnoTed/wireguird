package gui

// This file adds two features on top of the original wireguird:
//
//   1) The "Add Tunnel" button now opens a small menu with two choices:
//        - "Add empty tunnel…"   -> creates a brand new .conf (with a freshly
//                                    generated private key) and opens the editor
//        - "Import from file…"   -> the original browse-for-.conf/.zip behaviour
//
//   2) A live search box is placed on top of the left-hand tunnel list.
//      Typing filters the list as you type; pressing Esc clears the search.
//
// Everything lives in this single file so the changes to the original
// source are tiny (see INSTRUCTIONS.md). It is part of the same "gui"
// package, so it can freely use the existing helpers (window, TunnelsPath,
// createEditor, parseZipArchive, ScanTunnels, the get.* helpers, etc.).

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/UnnoTed/wireguird/gui/get"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// -----------------------------------------------------------------------------
// Feature 1: "Add Tunnel" menu (Add empty tunnel / Import from file)
// -----------------------------------------------------------------------------

// showAddTunnelMenu pops up a two-item menu at the "Add Tunnel" button.
func (t *Tunnels) showAddTunnelMenu(tl *gtk.ListBox, btn *gtk.Button) {
	menu, err := gtk.MenuNew()
	if err != nil {
		ShowError(window, err, "add tunnel menu error")
		return
	}

	// (a) create an empty tunnel in place
	miEmpty, err := gtk.MenuItemNewWithLabel("Add empty tunnel…")
	if err != nil {
		ShowError(window, err, "add tunnel menu error")
		return
	}
	miEmpty.Connect("activate", func() {
		t.createEmptyTunnel(tl)
	})
	menu.Append(miEmpty)

	// (b) import from file (the original "browse" behaviour)
	miImport, err := gtk.MenuItemNewWithLabel("Import from file… (.conf / .zip)")
	if err != nil {
		ShowError(window, err, "add tunnel menu error")
		return
	}
	miImport.Connect("activate", func() {
		t.importTunnelFiles()
	})
	menu.Append(miImport)

	menu.ShowAll()
	// Pops the menu at the mouse pointer, i.e. right on the button the user
	// just clicked. Passing nil is fine (no triggering event needed).
	menu.PopupAtPointer(nil)
}

// importTunnelFiles is the ORIGINAL "Add Tunnel" behaviour, moved verbatim into
// its own function: browse for one or more .conf / .zip files and import them
// into /etc/wireguard.
func (t *Tunnels) importTunnelFiles() {
	err := func() error {
		dialog, err := gtk.FileChooserNativeDialogNew("Wireguird - Choose tunnel files (*.conf, *.zip)", window, gtk.FILE_CHOOSER_ACTION_OPEN, "OK", "Cancel")
		if err != nil {
			return err
		}
		defer dialog.Destroy()

		// filter *.conf and *.zip files
		filter, err := gtk.FileFilterNew()
		if err != nil {
			return err
		}
		filter.AddPattern("*.conf")
		filter.AddPattern("*.zip")
		filter.SetName("*.conf / *.zip")
		dialog.AddFilter(filter)
		dialog.SetSelectMultiple(true)

		res := dialog.Run()
		if gtk.ResponseType(res) == gtk.RESPONSE_ACCEPT {
			list, err := dialog.GetFilenames()
			if err != nil {
				return err
			}

			for _, fname := range list {
				// if file is zip archive
				if strings.HasSuffix(fname, ".zip") {
					if err := parseZipArchive(fname); err != nil {
						return err
					}
					continue
				}

				data, err := ioutil.ReadFile(fname)
				if err != nil {
					return err
				}

				if err := ioutil.WriteFile(filepath.Join(TunnelsPath, filepath.Base(fname)), data, 0600); err != nil {
					return err
				}
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
}

// createEmptyTunnel asks for a name, writes a fresh /etc/wireguard/<name>.conf
// with a newly generated private key, refreshes the list, selects the new row
// and opens the editor so the user can fill in the details.
func (t *Tunnels) createEmptyTunnel(tl *gtk.ListBox) {
	err := func() error {
		name, ok, err := promptTunnelName()
		if err != nil {
			return err
		}
		if !ok {
			return nil // user cancelled
		}

		name = strings.TrimSpace(name)
		if name == "" {
			name = "tunnel"
		}
		// keep the name usable as a filename / interface name
		name = strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(name)
		name = strings.TrimSuffix(name, ".conf")

		path := filepath.Join(TunnelsPath, name+".conf")
		if _, statErr := os.Stat(path); statErr == nil {
			return errors.New("A tunnel named '" + name + "' already exists.")
		}

		// generate a private key so the new tunnel is ready to edit
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return err
		}

		tmpl := "[Interface]\n" +
			"PrivateKey = " + key.String() + "\n" +
			"Address = \n" +
			"DNS = \n" +
			"\n" +
			"[Peer]\n" +
			"PublicKey = \n" +
			"AllowedIPs = \n" +
			"Endpoint = \n"

		if err := ioutil.WriteFile(path, []byte(tmpl), 0600); err != nil {
			return err
		}

		if err := t.ScanTunnels(); err != nil {
			return err
		}

		// select the freshly created row, then open the editor for it
		selectTunnelRow(tl, name)
		openEditorForTunnel(name)
		return nil
	}()

	if err != nil {
		ShowError(window, err, "create empty tunnel error")
	}
}

// promptTunnelName shows a tiny modal dialog asking for the new tunnel's name.
// Returns (name, true, nil) if the user confirmed, or (_, false, nil) if cancelled.
func promptTunnelName() (string, bool, error) {
	d, err := gtk.DialogNewWithButtons(
		"New tunnel name",
		window,
		gtk.DIALOG_MODAL,
		[]interface{}{"Cancel", gtk.RESPONSE_CANCEL},
		[]interface{}{"Create", gtk.RESPONSE_ACCEPT},
	)
	if err != nil {
		return "", false, err
	}
	defer d.Destroy()

	ca, err := d.GetContentArea()
	if err != nil {
		return "", false, err
	}

	entry, err := gtk.EntryNew()
	if err != nil {
		return "", false, err
	}
	entry.SetPlaceholderText("e.g. home, office, eu-server")
	entry.SetMarginTop(8)
	entry.SetMarginBottom(8)
	entry.SetMarginStart(8)
	entry.SetMarginEnd(8)
	entry.SetActivatesDefault(true) // Enter confirms
	ca.Add(entry)

	d.SetDefaultResponse(gtk.RESPONSE_ACCEPT)
	d.ShowAll()

	res := d.Run()
	if res != gtk.RESPONSE_ACCEPT {
		return "", false, nil
	}

	text, err := entry.GetText()
	if err != nil {
		return "", false, err
	}
	return text, true, nil
}

// selectTunnelRow finds the row whose name matches and selects/focuses it.
func selectTunnelRow(tl *gtk.ListBox, name string) {
	for i := 0; ; i++ {
		row := tl.GetRowAtIndex(i)
		if row == nil {
			return
		}
		n, err := row.GetName()
		if err == nil && n == name {
			tl.SelectRow(row)
			row.GrabFocus()
			return
		}
	}
}

// openEditorForTunnel opens the editor window for the given tunnel name,
// pre-filling it with the config file's contents (same low-budget syntax
// highlighting the original "Edit" button uses).
func openEditorForTunnel(name string) {
	err := func() error {
		ext := ""
		if !strings.HasSuffix(name, ".conf") {
			ext = ".conf"
		}
		path := filepath.Join(TunnelsPath, name+ext)

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
		ShowError(window, err, "open editor error")
	}
}

// -----------------------------------------------------------------------------
// Feature 2: live search box above the tunnel list
// -----------------------------------------------------------------------------

// tunnelSearchText is the current (lowercased) filter text.
var tunnelSearchText string

// setupTunnelSearch installs a GtkListBox filter and inserts a search entry
// directly above the tunnel list. Filtering happens as the user types; the
// filter automatically applies to rows re-created by ScanTunnels too.
func setupTunnelSearch(tl *gtk.ListBox) error {
	// Decide row visibility based on the current search text. A row's "name"
	// is the tunnel's config name (what the user sees in the list), so we match
	// against that.
	tl.SetFilterFunc(func(row *gtk.ListBoxRow) bool {
		if tunnelSearchText == "" {
			return true
		}
		name, err := row.GetName()
		if err != nil {
			return true
		}
		return strings.Contains(strings.ToLower(name), tunnelSearchText)
	})

	search, err := gtk.SearchEntryNew()
	if err != nil {
		return err
	}
	search.SetPlaceholderText("Search tunnels")
	search.SetMarginBottom(4)

	// Filter as the user types. "changed" fires immediately on every keystroke.
	search.Connect("changed", func() {
		text, _ := search.GetText()
		tunnelSearchText = strings.ToLower(strings.TrimSpace(text))
		tl.InvalidateFilter()
	})

	// Esc clears the search (GtkSearchEntry emits "stop-search" on Escape).
	search.Connect("stop-search", func() {
		search.SetText("")
		tunnelSearchText = ""
		tl.InvalidateFilter()
	})

	// Insert the search entry at the very top of the left panel (the vertical
	// GtkBox that holds the list and the add/delete/zip buttons).
	if box := firstBoxAncestor(tl); box != nil {
		box.PackStart(search, false, true, 0)
		box.ReorderChild(search, 0) // move it above the list
		search.ShowAll()
	} else {
		log.Error().Msg("wireguird_addons: could not find left panel box for search entry")
	}
	return nil
}

// firstBoxAncestor walks up the widget tree from start and returns the nearest
// ancestor that is a *gtk.Box (the left panel box, in our case).
func firstBoxAncestor(start gtk.IWidget) *gtk.Box {
	var cur gtk.IWidget = start
	for i := 0; i < 8; i++ {
		parent, err := cur.ToWidget().GetParent()
		if err != nil || parent == nil {
			return nil
		}
		if box, ok := parent.(*gtk.Box); ok {
			return box
		}
		cur = parent
	}
	return nil
}
