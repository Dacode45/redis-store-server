package storage

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

func DefaultRedisConn() (*redis.Client, *rejson.Handler) {
	rh := rejson.NewReJSONHandler()

	cli := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cmd := cli.Ping(context.Background())
	if cmd.Err() != nil {
		panic(cmd.Err().Error())
	}

	rh.SetGoRedisClient(cli)
	return cli, rh
}

func RedisConn(addr, password string) (*redis.Client, *rejson.Handler) {
	rh := rejson.NewReJSONHandler()

	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password})
	cmd := cli.Ping(context.Background())
	if cmd.Err() != nil {
		panic(cmd.Err().Error())
	}

	rh.SetGoRedisClient(cli)
	return cli, rh
}
