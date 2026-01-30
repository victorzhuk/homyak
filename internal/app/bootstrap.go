package app

import (
	"github.com/victorzhuk/homyak/internal/handlers/httpsrv"
	"github.com/victorzhuk/homyak/web"
)

//nolint:unparam // ignore unparam
func bootstrap(cfg *Config) ([]worker, []initializer, []finalizer) {
	static, err := web.NewStatic()
	if err != nil {
		panic(err)
	}
	hs := httpsrv.New(
		static,
		cfg.HTTP.Addr,
		cfg.HTTP.MaxHeaderSizeMb*(1<<20),
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.Feedback.FormURL,
	)

	return []worker{hs}, nil, nil
}
