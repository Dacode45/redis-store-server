package main

import (
	"log"
	"os"

	"github.com/Dacode45/redis-store/pkg/models"
	"github.com/Dacode45/redis-store/pkg/server"
	"github.com/Dacode45/redis-store/pkg/storage"
)

func main() {
	namespace := os.Getenv("REDIS_STORE_NAMESPACE")
	url := os.Getenv("REDIS_STORE_URL")
	redisURL := os.Getenv("REDIS_STORE_REDIS_URL")
	password := os.Getenv("REDIS_STORE_REDIS_PASSWORD")

	if namespace == "" {
		log.Panicln("REDIS_STORE_NAMESPACE required")
		os.Exit(1)
	}
	if url == "" {
		log.Panicln("REDIS_STORE_URL required")
		os.Exit(1)
	}
	if redisURL == "" {
		log.Panicln("REDIS_STORE_REDIS_URL required")
		os.Exit(1)
	}

	mngr := models.NewDataManager(namespace)
	cli, rh := storage.RedisConn(redisURL, password)
	store := storage.NewStorage(mngr, cli, rh)
	serve := server.NewServer(store)

	serve.ListenAndServe(url)
}
