package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	APIURL            string
	ServerToken       string
	HeartbeatInterval time.Duration
	PollInterval      time.Duration
	HTTPTimeout       time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		APIURL:            strings.TrimRight(os.Getenv("API_URL"), "/"),
		ServerToken:       os.Getenv("SERVER_TOKEN"),
		HeartbeatInterval: 30 * time.Second,
		PollInterval:      5 * time.Second,
		HTTPTimeout:       15 * time.Second,
	}

	if err := applyDurationEnv(&cfg.HeartbeatInterval, "HEARTBEAT_INTERVAL"); err != nil {
		return Config{}, err
	}
	if err := applyDurationEnv(&cfg.PollInterval, "POLL_INTERVAL"); err != nil {
		return Config{}, err
	}
	if err := applyDurationEnv(&cfg.HTTPTimeout, "HTTP_TIMEOUT"); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error
	if c.APIURL == "" {
		errs = append(errs, errors.New("API_URL is required"))
	}
	if c.ServerToken == "" {
		errs = append(errs, errors.New("SERVER_TOKEN is required"))
	}
	if c.HeartbeatInterval <= 0 {
		errs = append(errs, errors.New("heartbeat interval must be > 0"))
	}
	if c.PollInterval <= 0 {
		errs = append(errs, errors.New("poll interval must be > 0"))
	}
	return errors.Join(errs...)
}

func applyDurationEnv(dst *time.Duration, key string) error {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}
	// Accept Go duration ("30s") or integer seconds.
	if d, err := time.ParseDuration(raw); err == nil {
		*dst = d
		return nil
	}
	sec, err := strconv.Atoi(raw)
	if err != nil {
		return fmt.Errorf("%s: invalid duration %q", key, raw)
	}
	*dst = time.Duration(sec) * time.Second
	return nil
}
