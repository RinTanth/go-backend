package config

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/RinTanth/go-common/codec"
	env "github.com/caarlos0/env/v11"
)

// Config is a struct that contains all configuration for the application
// NOTE: struct name should be in lowercase and field name should be in uppercase
// you can group the configuration by adding new struct
// Example:
//
//	type Config struct {
//			...
//			GCP gcp  // no need to add tag `env` for struct here.
//	}
//
// then create gcp struct with tag `env` for each field
//
//	type gcp struct {
//		ProjectID string `env:"GCP_PROJECT_ID"`
//	}
//
// you can add field without grouping them by adding new field with tag `env`
// Example:
//
//	type Config struct {
//		...
//		AppName string `env:"APP_NAME"`
//	}
type Config struct {
	Server        Server
	AccessControl AccessControl
	Postgres      Postgres
	Header        Header
	JWT           JWT
	GoogleClient  GoogleClient
	Aesgcm        Aesgcm
	Hash          Hash
}

type Server struct {
	Hostname string `env:"HOSTNAME"`
	Port     string `env:"PORT,notEmpty"`
}

type AccessControl struct {
	AllowOrigin string `env:"ACCESS_CONTROL_ALLOW_ORIGIN"`
}

type Header struct {
	RefIDHeaderKey string `env:"REF_ID_HEADER_KEY,notEmpty"`
}

type JWT struct {
	Issuer      string        `env:"JWT_ISSUER,notEmpty"`
	Audience    string        `env:"JWT_AUDIENCE,notEmpty"`
	ExpDuration time.Duration `env:"JWT_EXP_DURATION,notEmpty"`
	PrivateKey  string        `env:"SECRET_JWT_PRIVATE_KEY,notEmpty"`
}

type Postgres struct {
	Host     string `env:"DB_HOST,notEmpty"`
	Port     string `env:"DB_PORT,notEmpty"`
	Name     string `env:"DB_NAME,notEmpty"`
	User     string `env:"SECRET_DB_USER,notEmpty"`
	Password string `env:"SECRET_DB_PASSWORD,notEmpty"`
}

type GoogleClient struct {
	VerifyTokenURL    string `env:"GOOGLE_OAUTH2_VERIFY_TOKEN,notEmpty"`
	GetUserProfileURL string `env:"GOOGLE_OAUTH2_GET_USER_PROFILE,notEmpty"`
	RevokeTokenURL    string `env:"GOOGLE_OAUTH2_REVOKE_TOKEN,notEmpty"`
}

type Aesgcm struct {
	Key string `env:"SECRET_AESGCM_KEY,notEmpty"`
}

type Hash struct {
	Pepper string `env:"SECRET_HASH_PEPPER,notEmpty"`
}

var once sync.Once
var config Config

func prefix(e string) string {
	if e == "" {
		return ""
	}

	return fmt.Sprintf("%s_", e)
}

func C(envPrefix string) Config {
	once.Do(func() {
		opts := env.Options{
			Prefix: prefix(envPrefix),
		}

		var err error
		config, err = parseEnv[Config](opts)
		if err != nil {
			log.Fatal(err)
		}

		base64Coder := codec.NewBase64Coder()
		rawJWTPrivateKey, err := base64Coder.DecodeBase64(config.JWT.PrivateKey)
		if err != nil {
			log.Fatal(err)
		}
		config.JWT.PrivateKey = rawJWTPrivateKey

	})

	return config
}

// TODO: read config from xxx.yaml file that contains ${ENV} variable e.g. serviceDLTUrl: ${SERVICE_CORE_DLT_ACCOUNT_URL}
