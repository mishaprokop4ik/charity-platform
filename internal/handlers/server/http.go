package server

import (
	zlog "Kurajj/pkg/logger"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HTTP struct {
	server *http.Server
}

type TLSCertPair struct {
	Key  string
	Cert string
}

func NewHTTPServer(port int, certPaths TLSCertPair, h http.Handler) (*HTTP, error) {
	cert, err := tls.LoadX509KeyPair(certPaths.Cert, certPaths.Key)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		},
		Certificates: []tls.Certificate{
			cert,
		},
	}
	return &HTTP{
		server: &http.Server{
			Addr:           fmt.Sprintf(":%d", port),
			Handler:        h,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   60 * time.Second,
			IdleTimeout:    10 * time.Second,
			MaxHeaderBytes: 0,
			TLSConfig:      cfg,
			TLSNextProto:   make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		},
	}, nil
}

func (h *HTTP) Run() {
	go func() {
		if err := h.server.ListenAndServeTLS("",
			""); err != nil {
			zlog.Log.Error(err, "can not start https server")
			return
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	signal.Notify(sc, syscall.SIGTERM)
	sig := <-sc
	zlog.Log.Info("caught system", "signal", sig)
	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = h.server.Shutdown(tc)
	cancel()
	zlog.Log.WithName("storage").Info("server stopped", "time", time.Now().String())
}
