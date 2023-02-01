package main

import (
	"Kurajj/internal/config"
	handlers2 "Kurajj/internal/handlers"
	"Kurajj/internal/handlers/server"
	"Kurajj/internal/repository"
	logic "Kurajj/internal/services"
	zlog "Kurajj/pkg/logger"
	"flag"
	"os"
)

var (
	dbConfig    = flag.String("db-config", "configs/db.yaml", "Provide Database's config values")
	authConfig  = flag.String("auth-config", "configs/auth.yaml", "Provide Authentication's config values")
	emailConfig = flag.String("admin-config", "configs/gmail.yaml", "Provide Email's config values")
)

var port = flag.Int("port", 8080, "HTTP server port number")

func main() {
	flag.Parse()
	zlog.Init()

	dbConfig, err := config.NewDBConfigFromFile(*dbConfig)
	if err != nil {
		zlog.Log.Error(err, "could not read database config")
		os.Exit(1)
	}

	authConfig, err := config.NewAuthenticationConfigFromFile(*authConfig)
	if err != nil {
		zlog.Log.Error(err, "could not read authentication config")
		os.Exit(1)
	}

	emailConfig, err := config.NewEmailConfigFromFile(*emailConfig)
	if err != nil {
		zlog.Log.Error(err, "could not read email config")
		os.Exit(1)
	}

	conn, err := repository.NewConnector(dbConfig)
	if err != nil {
		zlog.Log.Error(err, "could not create connector")
		os.Exit(1)
	}
	repo := repository.New(conn)
	service := logic.New(repo, &authConfig, &emailConfig)
	handlers := handlers2.New(service)

	httpServer, err := server.NewHTTPServer(*port, server.TLSCertPair{
		Key:  "tls/key.pem",
		Cert: "tls/cert.pem",
	}, handlers.InitRoutes())
	httpServer.Run()
}
