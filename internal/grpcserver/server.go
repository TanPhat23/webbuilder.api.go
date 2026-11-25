package grpcserver

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const DefaultBindAddr = "127.0.0.1"

// Config controls bind address, authentication, TLS, and reflection for the gRPC server.
type Config struct {
	BindAddr         string
	Port             string
	AuthToken        string
	TLSCertFile      string
	TLSKeyFile       string
	EnableReflection bool
	Env              string
}

// LoadConfigFromEnv reads gRPC server configuration from environment variables
// and validates production security invariants.
func LoadConfigFromEnv(defaultPort string) (Config, error) {
	bind := strings.TrimSpace(os.Getenv("GRPC_BIND"))
	if bind == "" {
		bind = DefaultBindAddr
	}
	port := strings.TrimSpace(os.Getenv("GRPC_PORT"))
	if port == "" {
		port = strings.TrimSpace(os.Getenv("PORT"))
	}
	if port == "" {
		port = defaultPort
	}
	if port == "" {
		port = "8081"
	}

	env := strings.TrimSpace(os.Getenv("ENV"))
	authToken := os.Getenv("GRPC_AUTH_TOKEN")
	certFile := strings.TrimSpace(os.Getenv("GRPC_TLS_CERT"))
	keyFile := strings.TrimSpace(os.Getenv("GRPC_TLS_KEY"))
	reflectEnabled := strings.EqualFold(strings.TrimSpace(os.Getenv("GRPC_REFLECTION")), "true") ||
		strings.EqualFold(env, "development")

	cfg := Config{
		BindAddr:         bind,
		Port:             port,
		AuthToken:        authToken,
		TLSCertFile:      certFile,
		TLSKeyFile:       keyFile,
		EnableReflection: reflectEnabled,
		Env:              env,
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate enforces TLS and token rules for production non-loopback bindings.
func (c Config) Validate() error {
	hasCert := c.TLSCertFile != ""
	hasKey := c.TLSKeyFile != ""
	if hasCert != hasKey {
		return errors.New("both GRPC_TLS_CERT and GRPC_TLS_KEY must be set together")
	}

	if isProductionEnv(c.Env) && !IsLoopbackHost(c.BindAddr) {
		if !hasCert || !hasKey {
			return errors.New("gRPC non-loopback bind in production requires GRPC_TLS_CERT and GRPC_TLS_KEY")
		}
		if c.AuthToken == "" {
			return errors.New("gRPC non-loopback bind in production requires GRPC_AUTH_TOKEN")
		}
	}
	return nil
}

// Address returns the host:port listen address.
func (c Config) Address() string {
	return net.JoinHostPort(c.BindAddr, c.Port)
}

// NewServer builds a configured *grpc.Server with optional TLS and bearer token interceptors.
func NewServer(cfg Config) (*grpc.Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var opts []grpc.ServerOption
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load gRPC TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	if cfg.AuthToken != "" {
		opts = append(opts,
			grpc.UnaryInterceptor(UnaryAuthInterceptor(cfg.AuthToken)),
			grpc.StreamInterceptor(StreamAuthInterceptor(cfg.AuthToken)),
		)
	}

	srv := grpc.NewServer(opts...)
	if cfg.EnableReflection {
		reflection.Register(srv)
	}
	return srv, nil
}

// UnaryAuthInterceptor enforces Bearer token authentication on unary RPCs.
func UnaryAuthInterceptor(expectedToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := authorizeContext(ctx, expectedToken); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamAuthInterceptor enforces Bearer token authentication on streaming RPCs.
func StreamAuthInterceptor(expectedToken string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authorizeContext(ss.Context(), expectedToken); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func authorizeContext(ctx context.Context, expectedToken string) error {
	if expectedToken == "" {
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md.Get("authorization")
	for _, raw := range values {
		token, hasBearer := strings.CutPrefix(raw, "Bearer ")
		if !hasBearer {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1 {
			return nil
		}
	}
	return status.Error(codes.Unauthenticated, "invalid authorization token")
}

// IsLoopbackHost reports whether host resolves strictly to a loopback literal or localhost.
func IsLoopbackHost(host string) bool {
	trimmed := strings.TrimSpace(strings.Trim(host, "[]"))
	if strings.EqualFold(trimmed, "localhost") {
		return true
	}
	ip := net.ParseIP(trimmed)
	return ip != nil && ip.IsLoopback()
}

func isProductionEnv(env string) bool {
	return strings.EqualFold(env, "production") || strings.EqualFold(env, "prod")
}
