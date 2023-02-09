package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/MicahParks/keyfunc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
)

const (
	envRedisPassword      = "REDIS_PASSWORD"
	defaultRouterBasePath = "/redis"
	defaultRedisKeyPrefix = "kave:"
	jwksUrlFormat         = "https://%s/.well-known/jwks.json"
)

// Config holds application configuration
type Config struct {
	Address        string  `toml:"address"`
	RedisAddress   string  `toml:"redis_address"`
	RouterBasePath string  `toml:"router_base_path"`
	RedisKeyPrefix *string `toml:"redis_key_prefix"`
	TimeoutMs      int     `toml:"timeout_ms"`
	Auth           struct {
		Enabled bool   `toml:"enabled"`
		Domain  string `toml:"domain"`
	} `toml:"auth"`
}

func main() {
	// Read configuration from a TOML file
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic(err)
	}

	// Set base path
	routerBasePath := config.RouterBasePath
	if routerBasePath == "" {
		routerBasePath = defaultRouterBasePath
	}

	// Set redis key prefix
	redisKeyPrefix := defaultRedisKeyPrefix
	if config.RedisKeyPrefix != nil {
		redisKeyPrefix = *config.RedisKeyPrefix
	}

	// Get redis password from environment
	redisPassword := os.Getenv(envRedisPassword)

	// Set requests timeout
	timeout := time.Duration(config.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	// create a default context
	ctx := context.Background()

	// Connect to Redis
	client, err := NewRedisClient(ctx, &redis.Options{
		Addr:     config.RedisAddress,
		Password: redisPassword,
	})
	if err != nil {
		panic(err)
	}

	// Create a new KeyValue kvHandler
	kvHandler := NewKeyValueHandler(client, redisKeyPrefix, readKeyFromCtx)

	// Create a new router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)

	// add auth middleware if enabled
	if config.Auth.Enabled {
		authMiddleware := createAuthMiddleware(config.Auth.Domain)
		router.Use(authMiddleware.Handler)
	}

	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(timeout))

	// Add health route
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Printf("error writing response: %v", err)
		}
	})

	// Add redis routes
	router.Route(routerBasePath, func(r chi.Router) {
		r.Route("/{key}", func(r chi.Router) {
			// Add redis key to context
			r.Use(injectKeyInCtx)

			// Add permission check middleware
			if config.Auth.Enabled {
				permissionHandler := NewPermissionMiddleware(redisKeyPrefix, readKeyFromCtx, readPermissionsFromCtx)
				r.Use(permissionHandler.Handler)
			}

			r.Get("/", kvHandler.Get)
			r.Post("/", kvHandler.Set)
		})
	})

	// Start the server
	if err := http.ListenAndServe(config.Address, router); err != nil {
		log.Fatal(err)
	}
}

func createAuthMiddleware(domain string) AuthMiddleware {
	urlStr := fmt.Sprintf(jwksUrlFormat, domain)

	resp, err := http.DefaultClient.Get(urlStr)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	jwks, err := keyfunc.NewJSON(json.RawMessage(buf))
	if err != nil {
		log.Fatal(err)
	}

	parse := func(token string) ([]string, error) {
		options := jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()})

		claims := &claimsWithPermissions{}
		_, err := jwt.ParseWithClaims(token, claims, jwks.Keyfunc, options)
		return claims.Permissions, err
	}

	return NewAuthMiddleware(parse, writePermissionsToCtx)
}

func injectKeyInCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the key is valid
		key := chi.URLParam(r, "key")
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newContext := writeKeyToCtx(r.Context(), key)

		next.ServeHTTP(w, r.WithContext(newContext))
	})
}
