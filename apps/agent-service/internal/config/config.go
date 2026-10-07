package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultZMQURL     = "ipc:///tmp/skydock-agent.sock"
	DefaultAPIURL     = "http://localhost:3000/api/v1"
	DefaultWebURL     = "http://localhost:5173"
	DefaultAPITimeout = 30 * time.Second
)

// Config is process configuration loaded from the environment and .env.
type Config struct {
	WatchDirs  []string
	DBPath     string
	ZMQURL     string
	APIBaseURL string
	WebOrigin  string
	APITimeout time.Duration
}

var current Config

// Current returns the configuration from the last successful Load.
func Current() Config {
	return current
}

// Load reads .env when present, then parses configuration from the environment.
// Existing environment variables win over values in the file.
func Load() (Config, error) {
	if err := loadDotEnv(); err != nil {
		return Config{}, err
	}

	cfg, err := fromEnv()
	if err != nil {
		return Config{}, err
	}

	current = cfg
	return cfg, nil
}

func loadDotEnv() error {
	for _, name := range []string{".env", filepath.Join("apps", "agent-service", ".env")} {
		info, err := os.Stat(name)
		if err != nil || info.IsDir() {
			continue
		}
		if err := godotenv.Load(name); err != nil {
			return fmt.Errorf("load %s: %w", name, err)
		}
		return nil
	}
	return nil
}

func fromEnv() (Config, error) {
	watchDirs, err := watchDirs()
	if err != nil {
		return Config{}, err
	}

	apiBaseURL, err := apiBaseURL()
	if err != nil {
		return Config{}, err
	}

	apiTimeout, err := apiTimeout()
	if err != nil {
		return Config{}, err
	}

	webOrigin, err := webOriginURL()
	if err != nil {
		return Config{}, err
	}

	return Config{
		WatchDirs:  watchDirs,
		DBPath:     strings.TrimSpace(os.Getenv("SKYDOCK_DB_PATH")),
		ZMQURL:     getenv("SKYDOCK_AGENT_ZMQ_URL", DefaultZMQURL),
		APIBaseURL: apiBaseURL,
		WebOrigin:  webOrigin,
		APITimeout: apiTimeout,
	}, nil
}

func webOriginURL() (string, error) {
	raw := getenv("SKYDOCK_WEB_URL", DefaultWebURL)
	if raw == "" {
		return "", fmt.Errorf("SKYDOCK_WEB_URL is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("SKYDOCK_WEB_URL is invalid: %q", raw)
	}
	return strings.TrimRight(raw, "/"), nil
}

func apiBaseURL() (string, error) {
	raw := getenv("SKYDOCK_API_URL", DefaultAPIURL)
	if raw == "" {
		return "", fmt.Errorf("SKYDOCK_API_URL is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("SKYDOCK_API_URL is invalid: %q", raw)
	}

	return raw, nil
}

func apiTimeout() (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv("SKYDOCK_API_TIMEOUT"))
	if raw == "" {
		return DefaultAPITimeout, nil
	}

	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout <= 0 {
		return 0, fmt.Errorf("SKYDOCK_API_TIMEOUT is invalid: %q", raw)
	}

	return timeout, nil
}

func watchDirs() ([]string, error) {
	raw := strings.TrimSpace(os.Getenv("SKYDOCK_WATCH_DIRS"))
	if raw == "" {
		return nil, fmt.Errorf("SKYDOCK_WATCH_DIRS is required (set it in .env)")
	}

	dirs := splitList(raw)
	if len(dirs) == 0 {
		return nil, fmt.Errorf("SKYDOCK_WATCH_DIRS is required (set it in .env)")
	}

	expanded := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		path, err := expandHome(dir)
		if err != nil {
			return nil, fmt.Errorf("SKYDOCK_WATCH_DIRS: %w", err)
		}
		expanded = append(expanded, path)
	}
	return expanded, nil
}

func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
