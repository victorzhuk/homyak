package app

import (
	"github.com/victorzhuk/homyak/internal/handlers/httpsrv"
	"github.com/victorzhuk/homyak/web"
)

//nolint:unparam // ignore unparam
func bootstrap(cfg *config) ([]worker, []initializer, []finalizer) {
	static, err := web.NewStatic()
	if err != nil {
		panic(err)
	}
	hs := httpsrv.New(
		static,
		cfg.APP.HTTP.Addr,
		cfg.APP.HTTP.MaxHeaderSizeMb,
		cfg.APP.HTTP.ReadTimeout,
		cfg.APP.HTTP.WriteTimeout,
		cfg.APP.Feedback.FormURL,
	)

	return []worker{hs}, nil, nil
}
