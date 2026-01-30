package app

import (
	"fmt"
	"io"
	"time"

	"github.com/caarlos0/env/v11"
)

var c Config

func MustParseConfig(_ string) *Config {
	if err := env.Parse(&c); err != nil {
		panic(err)
	}

	return &c
}

func PrintEnvsUsage(w io.Writer) {
	_, err := fmt.Fprintln(w, "Environment variables:")
	if err != nil {
		panic(err)
	}

	vars := []struct {
		name string
		def  string
		desc string
	}{
		{"APP_DEBUG", "false", "Enable debug mode"},
		{"APP_ENV", "local", "Environment (local, development, production)"},
		{"APP_HTTP_ADDR", ":8080", "HTTP server address"},
		{"APP_HTTP_MAX_HEADER_SIZE_MB", "1", "Max header size in MB"},
		{"APP_HTTP_READ_TIMEOUT", "3s", "Read timeout"},
		{"APP_HTTP_WRITE_TIMEOUT", "3s", "Write timeout"},
		{"APP_FEEDBACK_FORM_URL", "http://localhost:8080", "Feedback form URL"},
	}

	for _, v := range vars {
		_, err := fmt.Fprintf(w, "  %-30s default: %-20s %s\n", v.name, v.def, v.desc)
		if err != nil {
			panic(err)
		}
	}
}

// Config holds application configuration.
type Config struct {
	Debug    bool   `env:"APP_DEBUG" envDefault:"false"`
	Env      string `env:"APP_ENV" envDefault:"local"`
	HTTP     HTTPConfig
	Feedback FeedbackConfig
}

// HTTPConfig holds HTTP server configuration.
type HTTPConfig struct {
	Addr            string        `env:"APP_HTTP_ADDR" envDefault:":8080"`
	MaxHeaderSizeMb int           `env:"APP_HTTP_MAX_HEADER_SIZE_MB" envDefault:"1"`
	ReadTimeout     time.Duration `env:"APP_HTTP_READ_TIMEOUT" envDefault:"3s"`
	WriteTimeout    time.Duration `env:"APP_HTTP_WRITE_TIMEOUT" envDefault:"3s"`
}

// FeedbackConfig holds feedback form configuration.
type FeedbackConfig struct {
	FormURL string `env:"APP_FEEDBACK_FORM_URL" envDefault:"http://localhost:8080"`
}
