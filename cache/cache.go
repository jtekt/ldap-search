package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	gocache "github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

// Cache is a type alias for easier imports
type Cache = gocache.Cache

// New creates and returns a Redis-backed cache instance
func New(redisAddr string) *Cache {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[ERROR] Failed to connect to Redis at %s: %v", redisAddr, err)
		log.Println("[INFO] Running without cache.")
		return nil // Return nil if the connection fails
	}

	return gocache.New(&gocache.Options{
		Redis:      redisClient,
		LocalCache: gocache.NewTinyLFU(60*5, time.Minute), // In-memory LRU layer - Can be used for local caching
	})
}

func HashKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	return "ldap-cache:" + hex.EncodeToString(hash[:])
}
