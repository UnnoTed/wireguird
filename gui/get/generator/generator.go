package main

import (
	"flag"
	"path/filepath"
	"strings"

	"github.com/ungerik/go-dry"
)

func main() {
	list := []string{
		"ApplicationWindow",
		"ScrolledWindow",
		"AboutDialog",
		"ColorButton",
		"MenuButton",
		"HeaderBar",
		"MenuItem",
		"Notebook",
		"TextView",
		"Viewport",
		"ListBox",
		"Spinner",
		"Button",
		"Switch",
		"Window",
		"Entry",
		"Label",
		"Stack",
		"Box",
	}

	target := ""
	flag.StringVar(&target, "target", "", "-target=./gui")
	flag.Parse()

	if err := Generate(list, target); err != nil {
		panic(err)
	}
}

func Generate(list []string, target string) error {

	txt := `package get

import (
	"errors"

	"github.com/gotk3/gotk3/gtk"
)

var Builder *gtk.Builder
`
	for _, el := range list {
		vel := strings.ToLower(el)

		txt += `
func ` + el + `(name string) (*gtk.` + el + `, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	` + vel + `1, ok := obj.(*gtk.` + el + `)
	if !ok {
		return nil, errors.New("cant get *gtk.` + el + `: " + name)
	}

	return ` + vel + `1, nil
}
`
	}

	if err := dry.FileSetString(filepath.Join(target, "./gtk.go"), txt); err != nil {
		return err
	}

	return nil
}
