package grpcserver

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLoadConfigFromEnvDefaultsAndValidation(t *testing.T) {
	t.Setenv("GRPC_BIND", "")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("GRPC_AUTH_TOKEN", "")
	t.Setenv("GRPC_TLS_CERT", "")
	t.Setenv("GRPC_TLS_KEY", "")
	t.Setenv("GRPC_REFLECTION", "")

	cfg, err := LoadConfigFromEnv("8081")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BindAddr != "127.0.0.1" || cfg.Address() != "127.0.0.1:8081" {
		t.Fatalf("unexpected default bind: %+v", cfg)
	}
	if cfg.EnableReflection {
		t.Fatal("reflection must be disabled by default")
	}

	// Production non-loopback without TLS + token must fail
	t.Setenv("ENV", "production")
	t.Setenv("GRPC_BIND", "0.0.0.0")
	if _, err := LoadConfigFromEnv("8081"); err == nil {
		t.Fatal("expected error for non-loopback production gRPC without TLS/token")
	}

	// Prod shorthand with TLS but missing auth token must fail
	t.Setenv("ENV", "prod")
	t.Setenv("GRPC_TLS_CERT", "/tmp/cert.pem")
	t.Setenv("GRPC_TLS_KEY", "/tmp/key.pem")
	if _, err := LoadConfigFromEnv("8081"); err == nil {
		t.Fatal("expected error when GRPC_AUTH_TOKEN is missing in prod non-loopback")
	}

	// Prod non-loopback with TLS + token validates
	t.Setenv("GRPC_AUTH_TOKEN", "secret-token")
	if _, err := LoadConfigFromEnv("8081"); err != nil {
		t.Fatalf("expected valid prod config with TLS + token, got %v", err)
	}

	// Reflection allowed only when GRPC_REFLECTION=true or ENV=development
	t.Setenv("GRPC_BIND", "127.0.0.1")
	t.Setenv("GRPC_TLS_CERT", "")
	t.Setenv("GRPC_TLS_KEY", "")
	t.Setenv("ENV", "development")
	cfg, err = LoadConfigFromEnv("8081")
	if err != nil || !cfg.EnableReflection {
		t.Fatalf("expected reflection enabled in development: cfg=%+v err=%v", cfg, err)
	}

	t.Setenv("ENV", "staging")
	t.Setenv("GRPC_REFLECTION", "true")
	cfg, err = LoadConfigFromEnv("8081")
	if err != nil || !cfg.EnableReflection {
		t.Fatalf("expected reflection enabled with GRPC_REFLECTION=true: cfg=%+v err=%v", cfg, err)
	}
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context { return f.ctx }

func TestAuthInterceptors(t *testing.T) {
	const secret = "super-secret-grpc-token"
	unary := UnaryAuthInterceptor(secret)
	stream := StreamAuthInterceptor(secret)

	called := false
	okUnary := func(ctx context.Context, req any) (any, error) {
		called = true
		return "ok", nil
	}
	okStream := func(srv any, ss grpc.ServerStream) error {
		called = true
		return nil
	}

	// Missing metadata -> Unauthenticated
	_, err := unary(context.Background(), nil, &grpc.UnaryServerInfo{}, okUnary)
	if status.Code(err) != codes.Unauthenticated || called {
		t.Fatalf("expected Unauthenticated, got err=%v called=%v", err, called)
	}

	// Wrong token -> Unauthenticated
	badCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer wrong-token"))
	if err := stream(nil, &fakeServerStream{ctx: badCtx}, &grpc.StreamServerInfo{}, okStream); status.Code(err) != codes.Unauthenticated || called {
		t.Fatalf("expected Unauthenticated for stream, got err=%v called=%v", err, called)
	}

	// Missing Bearer prefix -> Unauthenticated
	noBearerCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", secret))
	if _, err := unary(noBearerCtx, nil, &grpc.UnaryServerInfo{}, okUnary); status.Code(err) != codes.Unauthenticated || called {
		t.Fatalf("expected Unauthenticated without Bearer prefix, got err=%v", err)
	}

	// Valid token -> succeeds
	goodCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+secret))
	res, err := unary(goodCtx, nil, &grpc.UnaryServerInfo{}, okUnary)
	if err != nil || res != "ok" || !called {
		t.Fatalf("expected unary success, got res=%v err=%v called=%v", res, err, called)
	}
	called = false
	if err := stream(nil, &fakeServerStream{ctx: goodCtx}, &grpc.StreamServerInfo{}, okStream); err != nil || !called {
		t.Fatalf("expected stream success, got err=%v called=%v", err, called)
	}
}
