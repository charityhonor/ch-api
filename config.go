package altrudos

import (
	"errors"
	"fmt"
	"os"

	errs "github.com/pkg/errors"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"

	"github.com/altrudos/api/pkg/justgiving"
)

var (
	ErrInvalidJGMode       = errors.New("invalid justgiving mode")
	ErrBlankConfigFilePath = errors.New("config file path is blank, check your env vars")
)

type Config struct {
	Postgres   string
	Port       int
	JustGiving JGConfig
	WebsiteUrl string // Frontend
	BaseUrl    string //backend
}

type JGConfig struct {
	AppId string
	Mode  justgiving.Mode
}

type Services struct {
	DB *sqlx.DB
	JG *justgiving.JustGiving
}

func MustGetConfigServices(confFile string) *Services {
	s, err := GetConfigServices(confFile)
	if err != nil {
		panic(err)
	}
	return s
}

func GetConfigServices(confFile string) (*Services, error) {
	c, err := ParseConfig(confFile)
	if err != nil {
		return nil, err
	}
	return c.Connect()
}

func MustGetConfig(confFile string) *Config {
	conf, err := ParseConfig(confFile)
	if err != nil {
		panic(err)
	}
	return conf
}

func ParseConfig(confFile string) (*Config, error) {
	if confFile == "" {
		return nil, ErrBlankConfigFilePath
	}
	f, err := os.Open(confFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var conf Config
	if _, err := toml.DecodeReader(f, &conf); err != nil {
		return nil, errs.Wrap(err, fmt.Sprintf("Filepath: '%s'", confFile))
	}
	return &conf, nil
}

func (c *Config) MustConnect() *Services {
	s, err := c.Connect()
	if err != nil {
		panic(err)
	}
	return s
}

func (c *Config) Connect() (*Services, error) {
	db, err := c.ConnectPostgres()
	if err != nil {
		return nil, err
	}
	jg, err := c.ConnectJG()
	if err != nil {
		return nil, err
	}
	return &Services{
		DB: db,
		JG: jg,
	}, nil
}

func (c *Config) MustConnectPostgres() *sqlx.DB {
	db, err := c.ConnectPostgres()
	if err != nil {
		panic(err)
	}
	return db
}

func (c *Config) ConnectPostgres() (*sqlx.DB, error) {
	return GetPostgresConnection(c.Postgres)
}

func (c *Config) ConnectJG() (*justgiving.JustGiving, error) {
	return c.JustGiving.Connect()
}

func (c *JGConfig) Connect() (*justgiving.JustGiving, error) {
	// Make sure mode is valid.
	validModes := []justgiving.Mode{justgiving.ModeProduction, justgiving.ModeStaging}
	var isValid bool
	for _, v := range validModes {
		if c.Mode == v {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, ErrInvalidJGMode
	}

	return &justgiving.JustGiving{
		AppId: c.AppId,
		Mode:  justgiving.Mode(c.Mode),
	}, nil
}
