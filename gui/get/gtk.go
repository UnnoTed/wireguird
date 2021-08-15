package get

import (
	"errors"

	"github.com/gotk3/gotk3/gtk"
)

var Builder *gtk.Builder

func ApplicationWindow(name string) (*gtk.ApplicationWindow, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	applicationwindow1, ok := obj.(*gtk.ApplicationWindow)
	if !ok {
		return nil, errors.New("cant get *gtk.ApplicationWindow: " + name)
	}

	return applicationwindow1, nil
}

func ScrolledWindow(name string) (*gtk.ScrolledWindow, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	scrolledwindow1, ok := obj.(*gtk.ScrolledWindow)
	if !ok {
		return nil, errors.New("cant get *gtk.ScrolledWindow: " + name)
	}

	return scrolledwindow1, nil
}

func AboutDialog(name string) (*gtk.AboutDialog, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	aboutdialog1, ok := obj.(*gtk.AboutDialog)
	if !ok {
		return nil, errors.New("cant get *gtk.AboutDialog: " + name)
	}

	return aboutdialog1, nil
}

func ColorButton(name string) (*gtk.ColorButton, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	colorbutton1, ok := obj.(*gtk.ColorButton)
	if !ok {
		return nil, errors.New("cant get *gtk.ColorButton: " + name)
	}

	return colorbutton1, nil
}

func MenuButton(name string) (*gtk.MenuButton, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	menubutton1, ok := obj.(*gtk.MenuButton)
	if !ok {
		return nil, errors.New("cant get *gtk.MenuButton: " + name)
	}

	return menubutton1, nil
}

func HeaderBar(name string) (*gtk.HeaderBar, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	headerbar1, ok := obj.(*gtk.HeaderBar)
	if !ok {
		return nil, errors.New("cant get *gtk.HeaderBar: " + name)
	}

	return headerbar1, nil
}

func MenuItem(name string) (*gtk.MenuItem, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	menuitem1, ok := obj.(*gtk.MenuItem)
	if !ok {
		return nil, errors.New("cant get *gtk.MenuItem: " + name)
	}

	return menuitem1, nil
}

func Notebook(name string) (*gtk.Notebook, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	notebook1, ok := obj.(*gtk.Notebook)
	if !ok {
		return nil, errors.New("cant get *gtk.Notebook: " + name)
	}

	return notebook1, nil
}

func TextView(name string) (*gtk.TextView, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	textview1, ok := obj.(*gtk.TextView)
	if !ok {
		return nil, errors.New("cant get *gtk.TextView: " + name)
	}

	return textview1, nil
}

func Viewport(name string) (*gtk.Viewport, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	viewport1, ok := obj.(*gtk.Viewport)
	if !ok {
		return nil, errors.New("cant get *gtk.Viewport: " + name)
	}

	return viewport1, nil
}

func ListBox(name string) (*gtk.ListBox, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	listbox1, ok := obj.(*gtk.ListBox)
	if !ok {
		return nil, errors.New("cant get *gtk.ListBox: " + name)
	}

	return listbox1, nil
}

func Spinner(name string) (*gtk.Spinner, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	spinner1, ok := obj.(*gtk.Spinner)
	if !ok {
		return nil, errors.New("cant get *gtk.Spinner: " + name)
	}

	return spinner1, nil
}

func Button(name string) (*gtk.Button, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	button1, ok := obj.(*gtk.Button)
	if !ok {
		return nil, errors.New("cant get *gtk.Button: " + name)
	}

	return button1, nil
}

func Switch(name string) (*gtk.Switch, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	switch1, ok := obj.(*gtk.Switch)
	if !ok {
		return nil, errors.New("cant get *gtk.Switch: " + name)
	}

	return switch1, nil
}

func Window(name string) (*gtk.Window, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	window1, ok := obj.(*gtk.Window)
	if !ok {
		return nil, errors.New("cant get *gtk.Window: " + name)
	}

	return window1, nil
}

func Entry(name string) (*gtk.Entry, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	entry1, ok := obj.(*gtk.Entry)
	if !ok {
		return nil, errors.New("cant get *gtk.Entry: " + name)
	}

	return entry1, nil
}

func Label(name string) (*gtk.Label, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	label1, ok := obj.(*gtk.Label)
	if !ok {
		return nil, errors.New("cant get *gtk.Label: " + name)
	}

	return label1, nil
}

func Stack(name string) (*gtk.Stack, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	stack1, ok := obj.(*gtk.Stack)
	if !ok {
		return nil, errors.New("cant get *gtk.Stack: " + name)
	}

	return stack1, nil
}

func Box(name string) (*gtk.Box, error) {
	obj, err := Builder.GetObject(name)
	if err != nil {
		return nil, err
	}

	box1, ok := obj.(*gtk.Box)
	if !ok {
		return nil, errors.New("cant get *gtk.Box: " + name)
	}

	return box1, nil
}
