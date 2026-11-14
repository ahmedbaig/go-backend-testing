package testkit

import (
	"testing"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-monitoring"
)

func AppOptions(t testing.TB, name string) (httpx.Options, *LogCapture) {
	t.Helper()
	log, logs := Logger(t)
	return httpx.Options{
		Name:        name,
		Environment: "testing",
		Logger:      log,
		Prober:      monitoring.NewProber(name, monitoring.WithoutScheduling()),
	}, logs
}
