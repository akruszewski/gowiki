package redis

import (
	"github.com/akruszewski/awiki/settings"
	"github.com/go-redis/redis"
)

func Connect() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     settings.RedisAddr,
		Password: settings.RedisPassword, // no password set
		DB:       settings.RedisDB,       // use default DB
	})
}
