package containers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/postgres"
	"github.com/testcontainers/redis"
)

// TestInfrastructure manages test containers
type TestInfrastructure struct {
	ctx       context.Context
	postgres  *postgres.Container
	redis     *redis.Container
	networks  []testcontainers.Network
}

// NewTestInfrastructure creates new infrastructure manager
func NewTestInfrastructure(ctx context.Context) *TestInfrastructure {
	return &TestInfrastructure{
		ctx: ctx,
	}
}

// Start starts all required containers
func (ti *TestInfrastructure) Start(ctx context.Context) error {
	fmt.Println("Starting test infrastructure...")
	
	// Create network for services
	network, err := testcontainers.NewNetwork(ctx, "csic-test-network")
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}
	ti.networks = append(ti.networks, network)
	
	// Start PostgreSQL
	pgCtr, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("csic_platform"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithImage("postgres:15-alpine"),
		testcontainers.WithNetwork(network),
		testcontainers.WithNetworkAliases("postgres"),
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres: %w", err)
	}
	ti.postgres = pgCtr
	fmt.Println("✓ PostgreSQL started")
	
	// Start Redis
	redisCtr, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithNetwork(network),
		testcontainers.WithNetworkAliases("redis"),
	)
	if err != nil {
		return fmt.Errorf("failed to start redis: %w", err)
	}
	ti.redis = redisCtr
	fmt.Println("✓ Redis started")
	
	// Set environment variables for services
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", ti.postgres.GetPort("5432/tcp"))
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", ti.redis.GetPort("6379/tcp"))
	
	fmt.Println("✓ Test infrastructure started successfully")
	return nil
}

// Cleanup stops and removes all containers
func (ti *TestInfrastructure) Cleanup(ctx context.Context) {
	fmt.Println("Cleaning up test infrastructure...")
	
	// Stop containers
	if ti.redis != nil {
		ti.redis.Terminate(ctx)
		fmt.Println("✓ Redis stopped")
	}
	if ti.postgres != nil {
		ti.postgres.Terminate(ctx)
		fmt.Println("✓ PostgreSQL stopped")
	}
	
	// Remove networks
	for _, net := range ti.networks {
		net.Remove(ctx)
	}
	
	fmt.Println("✓ Test infrastructure cleaned up")
}

// GetPostgresConnectionString returns the PostgreSQL connection string
func (ti *TestInfrastructure) GetPostgresConnectionString() string {
	if ti.postgres == nil {
		return ""
	}
	host, _ := ti.postgres.Host(ctx)
	port, _ := ti.postgres.MappedPort(ctx, "5432/tcp")
	return fmt.Sprintf("postgres://postgres:postgres@%s:%s/csic_platform?sslmode=disable", host, port)
}

// GetRedisConnectionString returns the Redis connection string
func (ti *TestInfrastructure) GetRedisConnectionString() string {
	if ti.redis == nil {
		return ""
	}
	host, _ := ti.redis.Host(ctx)
	port, _ := ti.redis.MappedPort(ctx, "6379/tcp")
	return fmt.Sprintf("%s:%s", host, port)
}
