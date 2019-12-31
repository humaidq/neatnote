package settings

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	WorkingDir string
	ConfigPath = "config.toml"
	Config     Configuration
)

type Configuration struct {
	SitePort        string
	DevMode         bool
	UniEmailDomain  string
	EmailAddress    string
	EmailPassword   string
	EmailSMTPServer string
	DBConfig        DatabaseConfiguration
}

type DBType int

const (
	MySQL = iota
	SQLite
)

type DatabaseConfiguration struct {
	Type     DBType
	Host     string
	Name     string
	User     string
	Password string
	Path     string
}

func newConfig() Configuration {
	return Configuration{
		SitePort:        "8080",
		DevMode:         false,
		UniEmailDomain:  "@hw.ac.uk",
		EmailAddress:    "noreply@example.com",
		EmailPassword:   "emailpasswordhere",
		EmailSMTPServer: "smtp.migadu.com:587",
		DBConfig: DatabaseConfiguration{
			Type:     MySQL,
			Host:     "localhost:3306",
			Name:     "notes",
			User:     "notes",
			Password: "passwordhere",
			Path:     "data.db",
		},
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	var err error
	WorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatal("Cannot get working directory! ", err)
	}
}
func LoadConfig() {
	var err error
	if _, err = toml.DecodeFile(WorkingDir+"/"+ConfigPath, &Config); err != nil {
		log.Printf("Cannot load config file. Error: %s", err)
		if os.IsNotExist(err) {
			log.Println("Generating new configuration file, as it doesn't exist")
			var err error

			buf := new(bytes.Buffer)
			if err = toml.NewEncoder(buf).Encode(newConfig()); err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile(ConfigPath, buf.Bytes(), 0600)
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
	}
}
