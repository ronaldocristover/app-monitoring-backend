package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port            string
	Env             string
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	MaxOpen  int
	MaxIdle  int
}

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func Load(path string) (*Config, error) {
	// Set defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("SHUTDOWN_TIMEOUT", "30s")
	viper.SetDefault("POSTGRES_MAX_OPEN", 25)
	viper.SetDefault("POSTGRES_MAX_IDLE", 5)
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	viper.SetDefault("CORS_ALLOW_CREDENTIALS", "true")
	viper.SetDefault("CORS_MAX_AGE", "86400s")

	viper.SetConfigFile(path)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No config file found (%s), using environment variables", path)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:            viper.GetString("PORT"),
			Env:             viper.GetString("ENV"),
			ShutdownTimeout: viper.GetDuration("SHUTDOWN_TIMEOUT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetString("POSTGRES_PORT"),
			User:     viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
			DBName:   viper.GetString("POSTGRES_DB"),
			MaxOpen:  viper.GetInt("POSTGRES_MAX_OPEN"),
			MaxIdle:  viper.GetInt("POSTGRES_MAX_IDLE"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			Expiry:        viper.GetDuration("JWT_EXPIRY"),
			RefreshExpiry: viper.GetDuration("JWT_REFRESH_EXPIRY"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   viper.GetStringSlice("CORS_ALLOWED_ORIGINS"),
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
			AllowCredentials: viper.GetBool("CORS_ALLOW_CREDENTIALS"),
			MaxAge:           viper.GetDuration("CORS_MAX_AGE"),
		},
	}

	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Database.Port == "" {
		cfg.Database.Port = "5432"
	}

	if cfg.Server.Env != "production" {
		log.Printf("Config: port=%s env=%s db=%s:%s/%s",
			cfg.Server.Port, cfg.Server.Env, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	} else {
		fmt.Printf("Config loaded: port=%s\n", cfg.Server.Port)
	}

	return cfg, nil
}
