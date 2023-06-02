package main

import (
	"Kurajj/configs"
	handlers2 "Kurajj/internal/handlers"
	"Kurajj/internal/handlers/server"
	"Kurajj/internal/repository"
	logic "Kurajj/internal/services"
	zlog "Kurajj/pkg/logger"
	"flag"
	"github.com/joho/godotenv"
	"os"
)

var (
	dbConfig    = flag.String("db-config", "configs/db.yaml", "Provide Database's config values")
	authConfig  = flag.String("auth-config", "configs/auth.yaml", "Provide Authentication's config values")
	emailConfig = flag.String("admin-config", "configs/gmail.yaml", "Provide Email's config values")
)

var port = flag.Int("port", 8080, "HTTP server port number")

var (
	privateCertPath = flag.String("private-cert-path", "certs/cert.pem", "Path to private TLS certificate")
	publicCertPath  = flag.String("public-cert-path", "certs/cert-key.pem", "Path to public TLS certificate")
)

// @title           Swagger Core Charity Platform
// @version         1.0
// @description     Kurajj Charity Platform

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  mykhailo.prokopchyk@nure.ua

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
func main() {
	flag.Parse()
	zlog.Init()

	dbConfig, err := configs.NewDBConfigFromFile(*dbConfig)
	if err != nil {
		zlog.Log.Error(err, "could not read database config")
		os.Exit(1)
	}

	authConfig, err := configs.NewAuthenticationConfigFromFile(*authConfig)
	if err != nil {
		zlog.Log.Error(err, "could not read authentication config")
		os.Exit(1)
	}

	emailConfig, err := configs.NewEmailConfigFromFile(*emailConfig)
	if err != nil {
		zlog.Log.Error(err, "could not read email config")
		os.Exit(1)
	}

	conn, err := repository.NewConnector(dbConfig)
	if err != nil {
		zlog.Log.Error(err, "could not create connector")
		os.Exit(1)
	}

	err = godotenv.Load()
	if err != nil {
		zlog.Log.Error(err, "could not get env values")
		os.Exit(1)
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	secretKey := os.Getenv("SECRET_KEY")
	accessKey := os.Getenv("ACCESS_KEY")
	region := os.Getenv("REGION")

	repo := repository.New(conn, repository.AWSConfig{
		AccessKey:       accessKey,
		SecretAccessKey: secretKey,
		Region:          region,
		BucketName:      s3Bucket,
	})

	service := logic.New(repo, &authConfig, &emailConfig)
	handlers := handlers2.New(service)

	httpServer, err := server.NewHTTPServer(*port, server.TLSCertPair{
		Key:  *publicCertPath,
		Cert: *privateCertPath,
	}, handlers.InitRoutes())
	httpServer.Run()
}
