package app

import (
	"fmt"
	"io"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var c config

func MustParseConfig(configPath string) *config {
	if configPath != "" {
		err := cleanenv.ReadConfig(configPath, &c)
		if err != nil {
			panic(err)
		}
	} else {
		err := cleanenv.ReadEnv(&c)
		if err != nil {
			panic(err)
		}
	}

	return &c
}

func PrintEnvsUsage(w io.Writer) {
	description, err := cleanenv.GetDescription(c, nil)
	if err != nil {
		panic(err)
	}

	_, err = fmt.Fprintln(w, description)
	if err != nil {
		panic(err)
	}
}

type config struct {
	APP appCfg `yaml:"app" env-prefix:"APP_"`
}

type appCfg struct {
	IsDebug bool   `yaml:"debug" env:"DEBUG" env-default:"false"`
	Env     string `yaml:"env" env:"ENV" env-default:"local"`

	HTTP     httpCfg     `yaml:"http" env-prefix:"HTTP_"`
	Feedback feedbackCfg `yaml:"feedback" env-prefix:"FEEDBACK_"`
}

type httpCfg struct {
	Addr            string        `yaml:"addr" env:"ADDR" env-default:":8080"`
	MaxHeaderSizeMb int           `yaml:"max_header_size_mb" env:"MAX_HEADER_SIZE_MB" env-default:"1"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"3s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"3s"`
}

type feedbackCfg struct {
	FormURL string `yaml:"form_url" env:"FORM_URL" env-default:"http://localhost:8080"`
}
