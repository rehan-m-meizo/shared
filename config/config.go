package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var once sync.Once

func LoadNonGin(env *string) {
	var envFile string

	switch *env {
	case "release":
		envFile = ".env.production"
		log.Println("Loaded .env.production for production environment")
	case "debug":
	default:
		workspaceRoot, err := findWorkspaceRoot()
		if err != nil {
			log.Fatal(err)
		}
		envFile = filepath.Join(workspaceRoot, ".env.development")
		fmt.Println("Loaded .env.development for production environment", envFile)
		log.Println("Loaded .env.development for debug environment")
	case "test":
		workspaceRoot, err := findWorkspaceRoot()
		if err != nil {
			log.Fatal(err)
		}
		envFile = filepath.Join(workspaceRoot, ".env.test")
		log.Println("Loaded .env.test for test environment")
	}

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Panicf("%s not found\n", envFile)
		}
	}

	err := os.Setenv("TZ", os.Getenv("TZ"))
	if err != nil {
		log.Println("Failed to set timezone for the application")
	}
}

func LoadEnvFile() {

	var envFile string
	switch gin.Mode() {
	case "release":
		envFile = ".env.production"
		log.Println("Loaded .env.production for production environment")
	case "debug":
		workspaceRoot, err := findWorkspaceRoot()
		if err != nil {
			log.Fatal(err)
		}
		envFile = filepath.Join(workspaceRoot, ".env.development")
		fmt.Println("Loaded .env.development for production environment", envFile)
		log.Println("Loaded .env.development for debug environment")
	case "test":
		workspaceRoot, err := findWorkspaceRoot()
		if err != nil {
			log.Fatal(err)
		}
		envFile = filepath.Join(workspaceRoot, ".env.test")
		log.Println("Loaded .env.test for test environment")
	}

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Panicf("%s not found\n", envFile)
		}
	}

	err := os.Setenv("TZ", os.Getenv("TZ"))
	if err != nil {
		log.Println("Failed to set timezone for the application")
	}
}

// Config load a specified .env.development.development variable
func Config(key string) string {
	once.Do(LoadEnvFile)
	return os.Getenv(key)
}

func findWorkspaceRoot() (string, error) {
	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil // Found the workspace root
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	return "", os.ErrNotExist
}
