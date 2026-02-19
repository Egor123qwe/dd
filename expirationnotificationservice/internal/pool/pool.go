package pool

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/config"
)

const (
	msgChanSize = 1000
)

type RedisClient struct {
	client *redis.Client
	pubsub *redis.PubSub
}

type RedisPool struct {
	clients []*RedisClient
	msgChan chan string
}

func NewRedisPool(cfg config.Config) *RedisPool {
	dbNumbers := cfg.Redis.DBNumbers
	pool := &RedisPool{
		clients: make([]*RedisClient, len(dbNumbers)),
		msgChan: make(chan string, msgChanSize),
	}

	redisOptions := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	for i, db := range dbNumbers {
		client := redis.NewClient(&redis.Options{
			Addr:     redisOptions,
			Password: cfg.Redis.Password,
			DB:       db,
		})

		_, err := client.Do(
			context.Background(),
			"CONFIG",
			"SET",
			"notify-keyspace-events",
			"Ex").
			Result()
		if err != nil {
			log.Fatal().Err(err).Msg("Can not set keyspace events")
		}

		event := fmt.Sprintf("__keyevent@%d__:expired", db)

		pool.clients[i] = &RedisClient{
			client: client,
			pubsub: client.PSubscribe(context.Background(), event),
		}
	}

	return pool
}

func (p *RedisPool) Start() {
	for _, client := range p.clients {
		go func(c *RedisClient) {
			ch := c.pubsub.Channel()
			for msg := range ch {
				p.msgChan <- msg.Payload
			}
		}(client)
	}
}

func (p *RedisPool) GetMessages() <-chan string {
	return p.msgChan
}

func (p *RedisPool) Shutdown() {
	for _, client := range p.clients {
		client.pubsub.Close()
		client.client.Close()
	}
	
	close(p.msgChan)
}
