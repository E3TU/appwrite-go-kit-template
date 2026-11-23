// Package config parses the template's environment into a base.Config.
//
// This is the only place in the template that reads environment variables.
// The library itself never touches env, so a real app can replace this file
// with flags, a config file, or a secrets manager without touching handlers.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/e3tu/appwrite-base"
)

// Env is the parsed configuration plus the template-only settings.
type Env struct {
	Base base.Config

	// Listen is the HTTP listen address.
	Listen string
	// TodosTable is the table the todo handlers use.
	TodosTable string
}

// Load reads the environment.
//
// Required: APPWRITE_PROJECT_ID, APPWRITE_API_KEY, APPWRITE_ENDPOINT,
// APPWRITE_PUBLIC_ENDPOINT.
func Load() (Env, error) {
	cfg := base.Defaults()

	cfg.ProjectID = strings.TrimSpace(os.Getenv("APPWRITE_PROJECT_ID"))
	cfg.APIKey = strings.TrimSpace(os.Getenv("APPWRITE_API_KEY"))
	cfg.Endpoint = strings.TrimSpace(os.Getenv("APPWRITE_ENDPOINT"))
	cfg.PublicEndpoint = strings.TrimSpace(os.Getenv("APPWRITE_PUBLIC_ENDPOINT"))
	cfg.CookieName = strings.TrimSpace(os.Getenv("APPWRITE_COOKIE_NAME"))
	if cfg.CookieName == "" && cfg.ProjectID != "" {
		cfg.CookieName = "a_session_" + cfg.ProjectID
	}

	for name, value := range map[string]string{
		"APPWRITE_PROJECT_ID":      cfg.ProjectID,
		"APPWRITE_API_KEY":         cfg.APIKey,
		"APPWRITE_ENDPOINT":        cfg.Endpoint,
		"APPWRITE_PUBLIC_ENDPOINT": cfg.PublicEndpoint,
	} {
		if value == "" {
			return Env{}, fmt.Errorf("%s is required", name)
		}
	}

	if raw := strings.TrimSpace(os.Getenv("APPWRITE_TIMEOUT")); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return Env{}, fmt.Errorf("APPWRITE_TIMEOUT: %w", err)
		}
		cfg.Timeout = d
	}

	selfSigned, err := envBool("APPWRITE_SELF_SIGNED")
	if err != nil {
		return Env{}, err
	}
	cfg.SelfSigned = selfSigned

	authOn, err := envBool("APPWRITE_AUTH_ENABLED")
	if err != nil {
		return Env{}, err
	}
	if authOn {
		cfg.Auth = &base.AuthConfig{
			SessionHeader: strings.TrimSpace(os.Getenv("APPWRITE_SESSION_HEADER")),
		}
	}

	storageOn, err := envBool("APPWRITE_STORAGE_ENABLED")
	if err != nil {
		return Env{}, err
	}
	if storageOn {
		cfg.Storage = &base.StorageConfig{
			DefaultBucketID: strings.TrimSpace(os.Getenv("APPWRITE_STORAGE_BUCKET")),
		}
	}

	dbOn, err := envBool("APPWRITE_DB_ENABLED")
	if err != nil {
		return Env{}, err
	}
	todosTable := strings.TrimSpace(os.Getenv("APPWRITE_TODOS_TABLE"))
	if dbOn {
		cfg.DB = &base.DBConfig{
			DatabaseID: strings.TrimSpace(os.Getenv("APPWRITE_DATABASE_ID")),
		}
		if cfg.DB.DatabaseID == "" {
			return Env{}, fmt.Errorf("APPWRITE_DATABASE_ID is required when APPWRITE_DB_ENABLED is true")
		}
	}
	if todosTable == "" {
		todosTable = "todos"
	}

	listen := strings.TrimSpace(os.Getenv("APPWRITE_LISTEN"))
	if listen == "" {
		listen = ":8080"
	}

	return Env{
		Base:       cfg,
		Listen:     listen,
		TodosTable: todosTable,
	}, nil
}

// envBool reads an optional boolean. An unset variable is false.
func envBool(key string) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}
	return v, nil
}
