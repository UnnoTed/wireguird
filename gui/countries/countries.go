package countries

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"github.com/rs/zerolog/log"
)

const (
	mmdbURL  = "https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-country/ip-to-country.mmdb"
	mmdbFile = "./countries.mmdb"
)

var (
	initErr error
	once    sync.Once // OpenDatabase runs once
	db      *maxminddb.Reader

	ready      = make(chan struct{})
	dnsCache   = map[string]string{}
	dnsCacheMu sync.Mutex
)

type IPRecord struct {
	CountryCode string `maxminddb:"country_code"`
	Country     struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func OpenDatabase() error {
	once.Do(func() {
		defer close(ready)
		log.Debug().Msg("initializing country database")

		// downloads the db when doesnt exist or is older than a week
		if s, err := os.Stat(mmdbFile); os.IsNotExist(err) || s.ModTime().Before(time.Now().Add(-7*24*time.Hour)) {
			if err := downloadDatabase(); err != nil {
				initErr = fmt.Errorf("failed to download country database: %w", err)
				return
			}
		}

		var err error
		db, err = maxminddb.Open(mmdbFile)
		if err != nil {
			initErr = fmt.Errorf("failed to open country database: %w", err)
			return
		}

		log.Debug().Msg("country database opened successfully")
	})

	return initErr
}

func downloadDatabase() error {
	resp, err := http.Get(mmdbURL)
	if err != nil {
		return fmt.Errorf("failed to download mmdb: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(mmdbFile)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write data to disk: %w", err)
	}

	return nil
}

func Find(ipStr string) (string, error) {
	<-ready

	if initErr != nil {
		return "", fmt.Errorf("database initialization failed: %w", initErr)
	}

	if val, ok := dnsCache[ipStr]; ok {
		log.Debug().Str("ip", ipStr).Str("country", val).Msg("country found in cache")
		return val, nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address format: %s", ipStr)
	}

	var record IPRecord
	err := db.Lookup(ip, &record)
	if err != nil {
		return "", fmt.Errorf("error during database lookup: %w", err)
	}

	if record.CountryCode != "" {
		dnsCacheMu.Lock()
		dnsCache[ipStr] = record.CountryCode
		dnsCacheMu.Unlock()
		return record.CountryCode, nil
	}

	if record.Country.IsoCode != "" {
		dnsCacheMu.Lock()
		dnsCache[ipStr] = record.Country.IsoCode
		dnsCacheMu.Unlock()
		return record.Country.IsoCode, nil
	}

	return "Unknown", nil
}

func CloseDatabase() {
	if db != nil {
		db.Close()
	}
}
