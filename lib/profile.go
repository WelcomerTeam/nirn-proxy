package lib

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
)

func StartProfileServer() {
	slog.Info("Starting profiling server on :7654")

	http.ListenAndServe(":7654", nil)
}
