package app

import (
	"github.com/victorzhuk/homyak/internal/handlers/httpsrv"
	"github.com/victorzhuk/homyak/web"
)

func bootstrap(cfg *config) ([]worker, []initializer, []finalizer) {
	var (
		initializers []initializer
		finalizers   []finalizer
	)

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

	return []worker{hs}, initializers, finalizers
}
