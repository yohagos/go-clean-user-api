package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	Mode         string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type LogConfig struct {
	Level      string
	OutputPath string
	Encoding   string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     int
	RefreshTTL    int
}

func Load() (*Config, error) {
	viper.New()

	fmt.Println(
		"Starting config loader....",
	)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./..")

	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server.port", ":1111")
	viper.SetDefault("server.read_timeout", 15)
	viper.SetDefault("server.write_timeout", 15)
	viper.SetDefault("server.idle_timeout", 15)
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", 5)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.output_path", "stdout")
	viper.SetDefault("log.encoding", "json")

	viper.SetDefault("jwt.access_ttl", 1200)
	viper.SetDefault("jwt.refresh_ttl", 1209600)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err.Error())
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config

	cfg.Server.Port = viper.GetString("server.port")
	cfg.Server.ReadTimeout = viper.GetInt("server.read_timeout")
	cfg.Server.WriteTimeout = viper.GetInt("server.write_timeout")
	cfg.Server.IdleTimeout = viper.GetInt("server.idle_timeout")
	cfg.Server.Mode = viper.GetString("server.mode")

	dbURL := viper.GetString("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("ENV value for DATABASE_URL was empty. Will be set manualley")
		databaseURL := viper.GetString("database.url")
		if databaseURL == "" {
			fmt.Println(
				"database.url was empty and will be set manualley",
			)
			cfg.Database.URL = "postgres://postgres:postgres@postgres:5432/userdb?sslmode=disable"
		} else {
			cfg.Database.URL = databaseURL
		}
	} else {
		cfg.Database.URL = dbURL
	}

	cfg.Database.MaxOpenConns = viper.GetInt("database.max_open_conns")
	cfg.Database.MaxOpenConns = viper.GetInt("database.max_idle_conns")
	cfg.Database.MaxOpenConns = viper.GetInt("database.conn_max_lifetime")

	cfg.Log.Level = viper.GetString("log.level")
	cfg.Log.OutputPath = viper.GetString("log.output_path")
	cfg.Log.Encoding = viper.GetString("log.encoding")

	cfg.JWT.AccessSecret = viper.GetString("jwt.access_secret")
	cfg.JWT.RefreshSecret = viper.GetString("jwt.refresh_secret")
	cfg.JWT.AccessTTL = viper.GetInt("jwt.access_ttl")
	cfg.JWT.RefreshTTL = viper.GetInt("jwt.refresh_ttl")

	fmt.Println(
		"Config Load went through successfully",
	)

	return &cfg, nil
}
