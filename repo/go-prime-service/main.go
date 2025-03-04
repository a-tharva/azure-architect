package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strconv"

	pb "prime-service/proto" // Adjust path to your proto package

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

// Config holds the configuration parameters
type Config struct {
	RedisAddr     string
	RedisPassword string
	GRPCPort      string
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() Config {
	return Config{
		RedisAddr:     getEnv("REDIS_ADDR", "redis-cluster.redis.svc.cluster.local:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		GRPCPort:      getEnv("GRPC_PORT", "50051"),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

type server struct {
	pb.UnimplementedPrimeServiceServer
	rdb *redis.ClusterClient
}

func (s *server) GetPrimes(ctx context.Context, req *pb.PrimeRequest) (*pb.PrimeResponse, error) {
	ns := req.GetNs()
	if len(ns) == 0 {
		return &pb.PrimeResponse{}, nil
	}

	maxN := int32(0)
	for _, n := range ns {
		if n > maxN {
			maxN = n
		}
	}

	err := s.ensurePrimesUpTo(ctx, maxN)
	if err != nil {
		return nil, fmt.Errorf("failed to compute primes: %v", err)
	}

	var responsePrimes []int64
	for _, n := range ns {
		if n <= 0 {
			continue
		}
		primeStr, err := s.rdb.ZRange(ctx, "primes", int64(n-1), int64(n-1)).Result()
		if err != nil || len(primeStr) == 0 {
			return nil, fmt.Errorf("failed to retrieve prime for n=%d: %v", n, err)
		}
		prime, err := strconv.ParseInt(primeStr[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prime for n=%d: %v", n, err)
		}
		responsePrimes = append(responsePrimes, prime)
	}

	return &pb.PrimeResponse{Primes: responsePrimes}, nil
}

func (s *server) ensurePrimesUpTo(ctx context.Context, maxN int32) error {
	primeCount, err := s.rdb.ZCard(ctx, "primes").Result()
	if err != nil && err != redis.Nil {
		return err
	}

	if primeCount >= int64(maxN) {
		return nil
	}

	var primeList []int64
	if primeCount > 0 {
		primes, err := s.rdb.ZRangeWithScores(ctx, "primes", 0, -1).Result()
		if err != nil {
			return err
		}
		for _, z := range primes {
			p, ok := z.Member.(string)
			if !ok {
				return fmt.Errorf("invalid prime data in Redis")
			}
			primeNum, _ := strconv.ParseInt(p, 10, 64)
			primeList = append(primeList, primeNum)
		}
	}

	lastPrime := int64(1)
	if len(primeList) > 0 {
		lastPrime = primeList[len(primeList)-1]
	}
	candidate := lastPrime + 1

	for i := primeCount + 1; i <= int64(maxN); i++ {
		for {
			if isPrime(candidate, primeList) {
				break
			}
			candidate++
		}
		primeList = append(primeList, candidate)
		_, err := s.rdb.ZAdd(ctx, "primes", &redis.Z{
			Score:  float64(i),
			Member: candidate,
		}).Result()
		if err != nil {
			return err
		}
		candidate++
	}
	return nil
}

func isPrime(n int64, primes []int64) bool {
	sqrtN := int64(math.Sqrt(float64(n)))
	for _, p := range primes {
		if p > sqrtN {
			break
		}
		if n%p == 0 {
			return false
		}
	}
	return true
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize Redis client with configured values
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{config.RedisAddr},
		Password: config.RedisPassword,
	})

	// Listen on configured port
	lis, err := net.Listen("tcp", ":"+config.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", config.GRPCPort, err)
	}

	// Create and start gRPC server
	s := grpc.NewServer()
	pb.RegisterPrimeServiceServer(s, &server{rdb: rdb})
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
