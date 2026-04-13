package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
}

type AppConfig struct {
	Port string `yaml:"port"`
	Env  string `yaml:"env"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode"`
	Timezone string `yaml:"timezone"`
}

type JWTConfig struct {
	SecretKey      string `yaml:"secret_key"`
	AccessExpires  int    `yaml:"access_expires"`
	RefreshExpires int    `yaml:"refresh_expires"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.Name, d.Port, d.SSLMode, d.Timezone,
	)
}

func (j *JWTConfig) MaskSecret() string {
	if len(j.SecretKey) <= 8 {
		return "********"
	}
	return j.SecretKey[:4] + "********" + j.SecretKey[len(j.SecretKey)-4:]
}

func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("Config{\n")
	sb.WriteString(fmt.Sprintf("  App: {Port: %s, Env: %s},\n", c.App.Port, c.App.Env))
	sb.WriteString(fmt.Sprintf("  Database: {Host: %s, Port: %s, Name: %s, User: %s, Password: %s},\n",
		c.Database.Host, c.Database.Port, c.Database.Name, c.Database.User, "********"))
	sb.WriteString(fmt.Sprintf("  JWT: {SecretKey: %s, AccessExpires: %d, RefreshExpires: %d},\n",
		jwtConfig{SecretKey: c.JWT.MaskSecret()}.SecretKey, c.JWT.AccessExpires, c.JWT.RefreshExpires))
	sb.WriteString(fmt.Sprintf("  Log: {Level: %s, File: %s}\n", c.Log.Level, c.Log.File))
	sb.WriteString("}")
	return sb.String()
}

type jwtConfig struct {
	SecretKey string
}

var (
	cfg  *Config
	once sync.Once
	err  error
)

func Load() (*Config, error) {
	once.Do(func() {
		_ = godotenv.Load()

		env := getEnv("APP_ENV", "development")
		cfg, err = loadConfig(env)
	})
	return cfg, err
}

func Get() *Config {
	if cfg == nil {
		panic("config not loaded, call Load() first")
	}
	return cfg
}

func loadConfig(env string) (*Config, error) {
	basePath := getConfigPath("")

	cfg := &Config{}

	if err := mergeFile(cfg, basePath); err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	mergeEnv(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func getConfigPath(suffix string) string {
	dir := os.Getenv("CONFIG_DIR")
	if dir == "" {
		dir = "internal/config"
	}
	if suffix != "" {
		return fmt.Sprintf("%s/config%s.yaml", dir, suffix)
	}
	return fmt.Sprintf("%s/config.yaml", dir)
}

func mergeFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return nil
}

func mergeEnv(cfg *Config) {
	if v := getEnv("APP_PORT", ""); v != "" {
		cfg.App.Port = v
	}
	if v := getEnv("APP_ENV", ""); v != "" {
		cfg.App.Env = v
	}

	if v := getEnv("DB_HOST", ""); v != "" {
		cfg.Database.Host = v
	}
	if v := getEnv("DB_PORT", ""); v != "" {
		cfg.Database.Port = v
	}
	if v := getEnv("DB_NAME", ""); v != "" {
		cfg.Database.Name = v
	}
	if v := getEnv("DB_USER", ""); v != "" {
		cfg.Database.User = v
	}
	if v := getEnv("DB_PASSWORD", ""); v != "" {
		cfg.Database.Password = v
	}
	if v := getEnv("DB_SSLMODE", ""); v != "" {
		cfg.Database.SSLMode = v
	}

	if v := getEnv("JWT_SECRET_KEY", ""); v != "" {
		cfg.JWT.SecretKey = v
	}
	if v := getEnv("JWT_ACCESS_EXPIRES", ""); v != "" {
		if iv := parseIntEnv(v); iv > 0 {
			cfg.JWT.AccessExpires = iv
		}
	}
	if v := getEnv("JWT_REFRESH_EXPIRES", ""); v != "" {
		if iv := parseIntEnv(v); iv > 0 {
			cfg.JWT.RefreshExpires = iv
		}
	}

	if v := getEnv("LOG_LEVEL", ""); v != "" {
		cfg.Log.Level = v
	}
	if v := getEnv("LOG_FILE", ""); v != "" {
		cfg.Log.File = v
	}
}

func (c *Config) Validate() error {
	var errs []string

	if c.JWT.SecretKey == "" {
		errs = append(errs, "jwt.secret_key is required")
	} else if len(c.JWT.SecretKey) < 16 {
		errs = append(errs, "jwt.secret_key must be at least 16 characters")
	}

	if c.Database.Host == "" {
		errs = append(errs, "database.host is required")
	}

	if c.Database.Name == "" {
		errs = append(errs, "database.name is required")
	}

	if c.App.Port == "" {
		errs = append(errs, "app.port is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func parseIntEnv(v string) int {
	var result int
	fmt.Sscanf(v, "%d", &result)
	return result
}
