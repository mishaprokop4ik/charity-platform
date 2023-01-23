package config

import "fmt"

type DB struct {
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslMode"`
}

func (d DB) String() string {
	return fmt.Sprintf("host: %s, port: %d, database: %s, user: %s, password: %s, sslMode: %s",
		d.Host, d.Port, d.Database, d.User, d.Password, d.SSLMode)
}

func (d DB) DSN() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Europe/Kiev",
		d.User, d.Password, d.Database, d.Port, d.SSLMode)
}
