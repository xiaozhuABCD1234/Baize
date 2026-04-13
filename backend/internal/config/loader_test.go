package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func resetConfig() {
	cfg = nil
	once = sync.Once{}
	err = nil
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	absDir, _ := filepath.Abs(dir)

	yamlContent := `app:
  port: "3000"
  env: "test"

database:
  host: "localhost"
  port: "5432"
  name: "testdb"
  user: "testuser"
  password: "testpass"
  sslmode: "disable"
  timezone: "UTC"

jwt:
  secret_key: "test_secret_key_12345678"
  access_expires: 300
  refresh_expires: 3600

log:
  level: "debug"
  file: "test.log"
`
	configPath := filepath.Join(absDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	oldConfigDir := os.Getenv("CONFIG_DIR")
	oldAppEnv := os.Getenv("APP_ENV")
	defer func() {
		if oldConfigDir != "" {
			os.Setenv("CONFIG_DIR", oldConfigDir)
		} else {
			os.Unsetenv("CONFIG_DIR")
		}
		if oldAppEnv != "" {
			os.Setenv("APP_ENV", oldAppEnv)
		} else {
			os.Unsetenv("APP_ENV")
		}
	}()

	resetConfig()
	os.Setenv("CONFIG_DIR", absDir)
	os.Setenv("APP_ENV", "test")

	testCfg, loadErr := Load()
	if loadErr != nil {
		t.Fatalf("Load() error = %v", loadErr)
	}

	if testCfg.App.Port != "3000" {
		t.Errorf("App.Port = %s, want 3000", testCfg.App.Port)
	}
	if testCfg.App.Env != "test" {
		t.Errorf("App.Env = %s, want test", testCfg.App.Env)
	}
	if testCfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %s, want localhost", testCfg.Database.Host)
	}
	if testCfg.Database.Name != "testdb" {
		t.Errorf("Database.Name = %s, want testdb", testCfg.Database.Name)
	}
	if testCfg.JWT.SecretKey != "test_secret_key_12345678" {
		t.Errorf("JWT.SecretKey = %s, want test_secret_key_12345678", testCfg.JWT.SecretKey)
	}
}

func TestEnvOverride(t *testing.T) {
	dir := t.TempDir()
	absDir, _ := filepath.Abs(dir)

	yamlContent := `app:
  port: "1323"
  env: "development"

database:
  host: "localhost"
  port: "5432"
  name: "default_db"
  user: "default_user"
  password: ""
  sslmode: "disable"
  timezone: "UTC"

jwt:
  secret_key: "yaml_secret_key_12345678"
  access_expires: 900
  refresh_expires: 604800

log:
  level: "info"
  file: "log/server.log"
`
	basePath := filepath.Join(absDir, "config.yaml")
	if err := os.WriteFile(basePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	oldConfigDir := os.Getenv("CONFIG_DIR")
	oldAppEnv := os.Getenv("APP_ENV")
	oldDBHost := os.Getenv("DB_HOST")
	oldDBName := os.Getenv("DB_NAME")
	oldJWTSecret := os.Getenv("JWT_SECRET_KEY")
	defer func() {
		restore := func(key, val string) {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
		restore("CONFIG_DIR", oldConfigDir)
		restore("APP_ENV", oldAppEnv)
		restore("DB_HOST", oldDBHost)
		restore("DB_NAME", oldDBName)
		restore("JWT_SECRET_KEY", oldJWTSecret)
	}()

	resetConfig()
	os.Setenv("CONFIG_DIR", absDir)
	os.Setenv("APP_ENV", "development")
	os.Setenv("DB_HOST", "env_host")
	os.Setenv("DB_NAME", "env_db")
	os.Setenv("JWT_SECRET_KEY", "env_secret_key_12345")

	testCfg, loadErr := Load()
	if loadErr != nil {
		t.Fatalf("Load() error = %v", loadErr)
	}

	if testCfg.Database.Host != "env_host" {
		t.Errorf("Database.Host = %s, want env_host (env override)", testCfg.Database.Host)
	}
	if testCfg.Database.Name != "env_db" {
		t.Errorf("Database.Name = %s, want env_db (env override)", testCfg.Database.Name)
	}
	if testCfg.JWT.SecretKey != "env_secret_key_12345" {
		t.Errorf("JWT.SecretKey = %s, want env_secret_key_12345 (env override)", testCfg.JWT.SecretKey)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				App:      AppConfig{Port: "1323", Env: "development"},
				Database: DatabaseConfig{Host: "localhost", Name: "db", User: "user", Password: "pass", Port: "5432", SSLMode: "disable"},
				JWT:      JWTConfig{SecretKey: "valid_secret_key_123", AccessExpires: 900, RefreshExpires: 604800},
				Log:      LogConfig{Level: "info", File: "log/server.log"},
			},
			wantErr: false,
		},
		{
			name: "missing jwt secret",
			cfg: Config{
				App:      AppConfig{Port: "1323", Env: "development"},
				Database: DatabaseConfig{Host: "localhost", Name: "db"},
				JWT:      JWTConfig{SecretKey: ""},
			},
			wantErr: true,
		},
		{
			name: "jwt secret too short",
			cfg: Config{
				App:      AppConfig{Port: "1323", Env: "development"},
				Database: DatabaseConfig{Host: "localhost", Name: "db"},
				JWT:      JWTConfig{SecretKey: "short"},
			},
			wantErr: true,
		},
		{
			name: "missing database host",
			cfg: Config{
				App:      AppConfig{Port: "1323", Env: "development"},
				Database: DatabaseConfig{Host: "", Name: "db"},
				JWT:      JWTConfig{SecretKey: "valid_secret_key_123"},
			},
			wantErr: true,
		},
		{
			name: "missing database name",
			cfg: Config{
				App:      AppConfig{Port: "1323", Env: "development"},
				Database: DatabaseConfig{Host: "localhost", Name: ""},
				JWT:      JWTConfig{SecretKey: "valid_secret_key_123"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDSN(t *testing.T) {
	db := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	expected := "host=localhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable TimeZone=UTC"
	if got := db.DSN(); got != expected {
		t.Errorf("DSN() = %s, want %s", got, expected)
	}
}

func TestMaskSecret(t *testing.T) {
	jwt := JWTConfig{SecretKey: "abcdefghijklmnop"}

	result := jwt.MaskSecret()
	if result != "abcd********mnop" {
		t.Errorf("MaskSecret() = %s, want abcd********mnop", result)
	}
}

func TestMaskSecretShort(t *testing.T) {
	jwt := JWTConfig{SecretKey: "short"}

	result := jwt.MaskSecret()
	if result != "********" {
		t.Errorf("MaskSecret() = %s, want ********", result)
	}
}
