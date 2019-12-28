package settings

import (
	"math/rand"
	"os"
	"time"
)

var (
	SitePort        string
	EmailAddress    string
	EmailPassword   string
	EmailSMTPServer string
	DBConfig        DatabaseConfiguration
	DevMode         bool
)

type DatabaseConfiguration struct {
	Host     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func LoadConfig() {
	rand.Seed(time.Now().UTC().UnixNano())
	// We will load all from ENV
	SitePort = os.Getenv("port")
	EmailAddress = os.Getenv("email_address")
	EmailPassword = os.Getenv("email_password")
	EmailSMTPServer = os.Getenv("email_smtp_server")
	if os.Getenv("dev_mode") == "1" {
		DevMode = true
	}

	DBConfig = DatabaseConfiguration{
		Host:     os.Getenv("db_host"),
		Name:     os.Getenv("db_name"),
		User:     os.Getenv("db_user"),
		Password: os.Getenv("db_password"),
		SSLMode:  os.Getenv("db_ssl_mode"),
	}
}
