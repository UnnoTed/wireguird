package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ungerik/go-dry"
)

type Settings struct {
	MultipleTunnels bool
	StartOnTray     bool
	CheckUpdates    bool
	TunnelsPath     string
	Debug           bool
}

var (
	multipleTunnels *bool
	checkUpdates    *bool
	debug           *bool
	tray            *bool
	filePath        string
)

func (s *Settings) Init() error {
	log.Debug().Msg("Settings init")

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	filePath = filepath.Join(filepath.Dir(exePath), "wireguird.settings")

	if err := s.Load(); err != nil {
		return err
	}

	if s.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return nil
}

func (s *Settings) Save() error {
	log.Debug().Msg("saving settings")
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filePath, data, 0666); err != nil {
		return err
	}

	log.Debug().Msg("saved settings")
	return nil
}

func (s *Settings) Load() error {
	log.Debug().Msg("loading settings")
	if !dry.FileExists(filePath) {
		log.Debug().Msg("settings file doesnt exist")
		return nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	settings := &Settings{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	*s = *settings

	log.Debug().Interface("settings", s).Msg("loaded settings")
	return nil
}
