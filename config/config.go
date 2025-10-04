package config

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
    Basic struct {
        Username string
        Password string
    }
    App struct {
        Port string
    }
    Database struct {
        Host     string
        Port     string
        User     string
        Password string
        Name     string
    }
    Defaults struct {
        UserRole string `yaml:"user_role"`
    } `yaml:"defaults"`
    Users []YAMLUser `yaml:"users"`
}

type YAMLUser struct {
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

// DSN builds a lib/pq DSN
func (a AppConfig) DSN() string {
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        a.Database.Host, a.Database.Port, a.Database.User, a.Database.Password, a.Database.Name)
}

// Load loads .env and config.yml
func Load() (AppConfig, error) {
    _ = godotenv.Load()
    var cfg AppConfig
    cfg.Basic.Username = getenv("BASIC_USN", "admin")
    cfg.Basic.Password = getenv("BASIC_PW", "admin")
    cfg.App.Port = getenv("APP_PORT", "8080")
    cfg.Database.Host = getenv("DB_HOST", "localhost")
    cfg.Database.Port = getenv("DB_PORT", "5432")
    cfg.Database.User = getenv("DB_USER", "postgres")
    cfg.Database.Password = getenv("DB_PASSWORD", "postgresaja")
    cfg.Database.Name = getenv("DB_NAME", "restapi_db")

    // Optional config.yml
    if b, err := os.ReadFile("config.yml"); err == nil {
        _ = yaml.Unmarshal(b, &cfg)
    }
    // Optional config.yaml
    if b, err := os.ReadFile("config.yaml"); err == nil {
        _ = yaml.Unmarshal(b, &cfg)
    }
    return cfg, nil
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

// RunMigrations creates tables if not exists and seeds sample data
func RunMigrations(db *sql.DB) error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS idols (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            "group_name" VARCHAR(100) NOT NULL,
            position VARCHAR(100) NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            created_by VARCHAR(64) NOT NULL DEFAULT 'system',
            updated_by VARCHAR(64) NOT NULL DEFAULT 'system',
            deleted_at TIMESTAMPTZ NULL,
            version INT NOT NULL DEFAULT 1
        );`,
        // app uses YAML users for authentication; DB users table is for listing
        // seed idols
        `INSERT INTO idols (name, "group_name", position)
         SELECT 'Jisung','NCT','Main Dancer'
         WHERE NOT EXISTS (SELECT 1 FROM idols WHERE name='Jisung');`,
        `INSERT INTO idols (name, "group_name", position)
         SELECT 'Karina','AESPA','Leader'
         WHERE NOT EXISTS (SELECT 1 FROM idols WHERE name='Karina');`,
    }
    for _, s := range stmts {
        if _, err := db.Exec(s); err != nil {
            return err
        }
    }
    return nil
}


