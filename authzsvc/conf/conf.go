package conf

import (
	"errors"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	ErrLoadConf          = errors.New("conf: error loading configuration")
	ErrReloadConf        = errors.New("conf: error reloading configuration")
	ErrAlreadyLoaded     = errors.New("conf: configuration already exist")
	ErrNoConfigFileFound = errors.New("conf: configuration file not provided/found")

	// C is the global configuration obj
	C = &Config{}

	defaults = map[string]interface{}{
		"auth": map[string]interface{}{
			"secret_key":        "topsecret",
			"issuer":            "Reactive Micro Org",
			"access_token_ttl":  time.Minute * 30,
			"refresh_token_ttl": time.Hour * 24,
			"access_kid":        "id_at",
			"refresh_kid":       "id_rt",
		},
	}
)

// Config hold the user service configuration
type Config struct {
	Env string `mapstructure:"env"`

	ReqIDKey string `mapstructure:"req_id_key"`
	SVCName  string `mapstructure:"svc_name"`
	MongoURI string `mapstructure:"mongo_uri"`
	NATSUrl  string `mapstructure:"nats_url"`

	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	Auth struct {
		SecretKey       string        `mapstructure:"secret_key"`
		Issuer          string        `mapstructure:"issuer"`
		AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
		RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
		AccessKID       string        `mapstructure:"access_kid"`
		RefreshKID      string        `mapstructure:"refresh_kid"`
	} `mapstructure:"auth"`
}

func (c *Config) Load(confFname string) error {
	v := viper.New()

	if *c != (Config{}) {
		return ErrAlreadyLoaded
	}

	if confFname != "" {
		v.SetConfigFile(confFname)
	} else {
		dir, _ := os.Getwd()
		dir = path.Join(dir, "conf")
		v.SetConfigName("conf")
		v.SetConfigType("json")
		v.AddConfigPath(dir)
	}

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return ErrNoConfigFileFound
	}

	// overwrite the configurations with the provided env variables
	v.SetEnvPrefix("authzsvc")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// var conf Config
	if err := v.Unmarshal(&c); err != nil {
		return ErrLoadConf
	}

	// set the global config obj
	C = c

	return nil
}

// Load reads the config file and returns read configs
func Load(confFname string) (*Config, error) {
	var conf Config
	err := conf.Load(confFname)
	return &conf, err
}

func (c *Config) GetURI() (dsn string) {
	return "mongodb://localhost:27017/authzdb?w=majority"
}

func setDefaults(v *viper.Viper) {
	for key, val := range defaults {
		v.SetDefault(key, val)
	}
}
