package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	JWT         JWTConfig
	App         AppConfig
	GoogleOAuth GoogleOAuthConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	TimeZone        string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type JWTConfig struct {
	SecretKey string
	Issuer    string
	ExpiresIn time.Duration
}

type AppConfig struct {
	Environment string
	LogLevel    string
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Enabled      bool
}

func Load() (*Config, error) {
	database, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}
	jwtCfg, err := loadJWTConfig()
	if err != nil {
		return nil, err
	}
	return &Config{
		Database:    database,
		Server:      loadServerConfig(),
		JWT:         jwtCfg,
		App:         loadAppConfig(),
		GoogleOAuth: loadGoogleOAuthConfig(),
	}, nil
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnvAsInt("DB_PORT", 5432)
	user, err := requireEnv("DB_USER")
	if err != nil {
		return DatabaseConfig{}, err
	}
	password, err := requireEnv("DB_PASSWORD")
	if err != nil {
		return DatabaseConfig{}, err
	}
	dbName := getEnv("DB_NAME", "postgres")
	sslMode := getEnv("DB_SSL_MODE", "disable")
	timeZone := getEnv("DB_TIMEZONE", "UTC")
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)

	dsn := buildDSN(host, port, user, password, dbName, sslMode, timeZone)

	return DatabaseConfig{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		DBName:          dbName,
		SSLMode:         sslMode,
		TimeZone:        timeZone,
		DSN:             dsn,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}, nil
}

func loadServerConfig() ServerConfig {
	port := getEnv("SERVER_PORT", "8080")
	readTimeout := getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second)
	writeTimeout := getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second)
	idleTimeout := getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second)

	return ServerConfig{
		Port:         port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

func loadJWTConfig() (JWTConfig, error) {
	secretKey, err := requireEnv("JWT_SECRET_KEY")
	if err != nil {
		return JWTConfig{}, err
	}
	if len(secretKey) < 32 {
		return JWTConfig{}, fmt.Errorf("JWT_SECRET_KEY deve possuir no mínimo 32 caracteres")
	}
	issuer := getEnv("JWT_ISSUER", "fynance_api")
	expiresIn := getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour)

	return JWTConfig{
		SecretKey: secretKey,
		Issuer:    issuer,
		ExpiresIn: expiresIn,
	}, nil
}

func loadAppConfig() AppConfig {
	environment := getEnv("APP_ENV", "development")
	logLevel := getEnv("LOG_LEVEL", "info")

	return AppConfig{
		Environment: environment,
		LogLevel:    logLevel,
	}
}

func buildDSN(host string, port int, user, password, dbName, sslMode, timeZone string) string {
	return "host=" + host +
		" user=" + user +
		" password=" + password +
		" dbname=" + dbName +
		" port=" + strconv.Itoa(port) +
		" sslmode=" + sslMode +
		" TimeZone=" + timeZone
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	if duration, err := time.ParseDuration(valueStr); err == nil {
		return duration
	}
	return defaultValue
}

func requireEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s precisa ser definido", key)
	}
	return value, nil
}

func loadGoogleOAuthConfig() GoogleOAuthConfig {
	clientID := getEnv("GOOGLE_OAUTH_CLIENT_ID", "")
	clientSecret := getEnv("GOOGLE_OAUTH_CLIENT_SECRET", "")
	redirectURL := getEnv("GOOGLE_OAUTH_REDIRECT_URL", "")
	enabledStr := strings.ToLower(strings.TrimSpace(getEnv("GOOGLE_OAUTH_ENABLED", "false")))
	enabled := enabledStr == "true" || enabledStr == "1"

	return GoogleOAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Enabled:      enabled,
	}
}
