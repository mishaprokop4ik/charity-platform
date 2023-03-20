package repository

import (
	"Kurajj/configs"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Connector struct {
	DB *gorm.DB
}

func NewConnector(config configs.DB) (*Connector, error) {
	conn := postgres.Open(config.DSN())
	db, err := gorm.Open(conn)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres err: %s\nconfig: %s", err, config)
	}

	return &Connector{DB: db}, nil
}
