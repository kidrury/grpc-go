package environment

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Environment string

const (
	EnvProduction  Environment = "production"
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
)

type Config struct {
	GRPCAddress     string
	Env             Environment
	DatabaseURL     string
	JWTSecret       string
	TLSEnabled      bool
	TLSCert         string
	TLSKey          string
	LogLevel        string
	ShutdownTimeout time.Duration
}

// so that I can use other sources for secrets as well
type SecretProvider interface {
	GetSecret(ctx context.Context, key string) (string, error)
}

// the source being env
type EnvSecretProvider struct{}

func (*EnvSecretProvider) GetSecret(ctx context.Context, key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("secret %q is not set", key)
	}

	return v, nil
}

var fallbackJWTSecret string = "only_for_development"

func Load(ctx context.Context, secret SecretProvider) (*Config, error) {
	err := loadDotEnvIfPresent(".env")

	if err != nil {
		return &Config{}, fmt.Errorf("loading .env: %w", err)
	}

	//loading from flags
	fs := flag.NewFlagSet("grpc-blog", flag.ContinueOnError)
	addrFlag := fs.String("addr", "", "override GRPC_ADDR (highest precedence)")
	envFlag := fs.String("addr", "", "override ENV (highest precedence)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return &Config{}, fmt.Errorf("parsing flags: %w", err)
	}

	//loading from env starts here
	//for variables: env and address, the values given by flags have priority
	env := Environment(getEnvOr("ENV", string(EnvDevelopment)))
	//replace by flag value if provided
	if *envFlag != "" {
		env = Environment(env)
	}

	addr, err := parseAddr("GRPC_ADDR", ":50051")
	if err != nil {
		return &Config{}, fmt.Errorf("loading address: %w", err)
	}
	//replace by flag value if provided
	if *addrFlag != "" {
		if err := validateAddr(*addrFlag); err != nil {
			return &Config{}, fmt.Errorf("validating flag address: %w", err)
		}
		addr = *addrFlag
	}

	databaseURL, err := getEnv("DATABASE_URL")
	if err != nil {
		return &Config{}, fmt.Errorf("loading database url: %w", err)
	}

	jwtSecret, err := secret.GetSecret(ctx, "JWT_SECRET")
	if err != nil {
		if env == EnvDevelopment {
			jwtSecret = fallbackJWTSecret
		} else {
			return &Config{}, fmt.Errorf("loading jwt secret: %w", err)
		}
	}

	tlsEnabled, err := parseBoolOr("TLS_ENABLED", false)
	if err != nil {
		return &Config{}, fmt.Errorf("loading tls enabled: %w", err)
	}

	tlsCert := getEnvOr("TLS_CERT", "server.crt")

	tlsKey := getEnvOr("TLS_KEY", "server.key")

	logLevel := getEnvOr("LOG_LEVEL", "info")

	shutdownTimeout, err := parseDuration("SHUTDOWN_TIMEOUT", time.Second*30)

	conf := &Config{
		GRPCAddress:     addr,
		Env:             env,
		DatabaseURL:     databaseURL,
		JWTSecret:       jwtSecret,
		TLSEnabled:      tlsEnabled,
		TLSCert:         tlsCert,
		TLSKey:          tlsKey,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
	}

	return conf, nil
}

func (c *Config) Validate() error {
	var errs []string

	switch c.Env {
	case EnvDevelopment, EnvProduction, EnvStaging:
	default:
		errs = append(errs, fmt.Sprintf("invalid ENV: %s - must be one of development|staging|production", c.Env))
	}

	//no need to check if JWTSecret is an empty string - would have already failed if that was the case
	if c.Env != EnvDevelopment && c.JWTSecret == fallbackJWTSecret {
		errs = append(errs, fmt.Sprintf("JWT_SECRET is the known development placeholder — refusing to start in %s", c.Env))
	}
	if c.Env != EnvDevelopment && len(c.JWTSecret) < 32 {
		errs = append(errs, fmt.Sprintf("JWT_SECRET must be at least 32 bytes in %s (got %d)", c.Env, len(c.JWTSecret)))
	}

	if c.TLSEnabled {
		if _, err := tls.LoadX509KeyPair(c.TLSCert, c.TLSKey); err != nil {
			errs = append(errs, fmt.Sprintf("failed to load TLS cert/key pair: %v", err))
		}
	}

	if c.Env == EnvProduction && !c.TLSEnabled {
		errs = append(errs, "TLS_ENABLED must be true in production")
	}

	if c.ShutdownTimeout <= 0 {
		errs = append(errs, "SHUTDOWN_TIMEOUT must be a positive duration")
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, fmt.Sprintf("LOG_LEVEL must be one of debug|info|warn|error, got %q", c.LogLevel))
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s):\n  - %s", len(errs), strings.Join(errs, "\n  - "))
	}
	return nil
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func getEnv(key string) (string, error) {
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("no value set for %s\n", key)
}

func parseAddr(key, fallback string) (string, error) {
	addr := os.Getenv(key)

	if addr == "" {
		if err := validateAddr(fallback); err != nil {
			return "", fmt.Errorf("validating fallback address: %w", err)
		}
		return fallback, nil
	}

	if err := validateAddr(addr); err != nil {
		return "", fmt.Errorf("validating env address: %w", err)
	}

	return addr, nil
}
func validateAddr(addr string) error {
	if !strings.HasPrefix(addr, ":") {
		return fmt.Errorf("%q must start with ':'", addr)
	}

	port := addr[1:]

	if port == "" {
		return fmt.Errorf("%q must contain a port after ':'", addr)
	}

	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("%q must be ':' followed by an integer", addr)
	}

	return nil
}

func parseBoolOr(key string, fallback bool) (bool, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}

	res, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("%s: %q is not a valid boolean", key, val)
	}
	return res, nil
}

func parseDuration(key string, fallback time.Duration) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}
	dur, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not a valid duration (e.g. \"30s\", \"1m\")", key, val)
	}

	return dur, nil
}

func loadDotEnvIfPresent(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for lineNum, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, found := strings.Cut(line, "=")

		if !found {
			return fmt.Errorf(
				"%s:%d expected format: KEY=VALUE , received %v",
				path,
				lineNum+1,
				line,
			)
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)

		if _, alreadySet := os.LookupEnv(key); !alreadySet {
			if err := os.Setenv(key, val); err != nil {
				return fmt.Errorf("setting %s:%v", val, err)
			}
		}
	}

	return nil
}
