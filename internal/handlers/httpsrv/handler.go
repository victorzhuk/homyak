package httpsrv

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/victorzhuk/homyak/internal/logger"
)

type srv struct {
	static fs.FS

	addr           string
	maxHeaderBytes int
	readTimeout    time.Duration
	writeTimeout   time.Duration

	feedbackFormURL string
}

func New(static fs.FS, addr string, maxHeaderSize int, readTimeout, writeTimeout time.Duration, feedbackFormURL string) *srv {
	return &srv{
		static:          static,
		addr:            addr,
		maxHeaderBytes:  maxHeaderSize,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		feedbackFormURL: feedbackFormURL,
	}
}

func (srv *srv) Start(ctx context.Context) error {
	logger.Debug(
		"creating http server",
		zap.String("addr", srv.addr),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.staticHandler)
	mux.HandleFunc("/feedback", srv.feedbackFormRedirect)

	ss := &http.Server{
		Handler:        mux,
		Addr:           srv.addr,
		MaxHeaderBytes: srv.maxHeaderBytes,
		ReadTimeout:    srv.readTimeout,
		WriteTimeout:   srv.writeTimeout,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down http handler")
		if err := ss.Shutdown(ctx); err != nil {
			logger.Error("error shutting down http handler", zap.Error(err))
		}
	}()

	logger.Info(
		"listening on http address",
		zap.String("addr", srv.addr),
	)
	err := ss.ListenAndServe()
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return fmt.Errorf("http server error: %w", err)
}
