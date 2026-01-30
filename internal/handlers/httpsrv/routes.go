package httpsrv

import (
	"net/http"
	"path"
	"strings"

	"go.uber.org/zap"

	"github.com/victorzhuk/homyak/internal/logger"
)

func (srv *srv) staticHandler(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	lwr := NewLoggingResponseWriter(w)

	http.ServeFileFS(lwr, r, srv.static, path.Clean(upath))
	logger.Debug(
		"access static",
		zap.String("path", r.URL.Path),
		zap.Int("status", lwr.statusCode),
	)
}

func (srv *srv) feedbackFormRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, srv.feedbackFormURL, http.StatusFound)
}

func (srv *srv) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//nolint:gosec,errcheck // Health check response write failure is non-critical
	w.Write([]byte("ok"))
}

func (srv *srv) readyzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//nolint:gosec,errcheck // Health check response write failure is non-critical
	w.Write([]byte("ready"))
}
