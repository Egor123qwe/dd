package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
)

func (c cache) SetOrUpdateMerchant(ctx context.Context, key string, msg message.MerchantRent) error {
	MerchantRent := model.Merchant{
		RequestID: msg.Content.RequestID,
		SessionID: msg.Content.SessionID,
		Settings:  msg.Content.Settings,
		ClientID:  msg.Content.ClientID,
	}

	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	val, err := json.Marshal(MerchantRent)
	if err != nil {
		return fmt.Errorf("error marshalling message : %v", err)
	}

	err = c.client.Watch(ctx, func(tx *redis.Tx) error {
		oldVal, err := tx.Get(ctx, key).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
					pipe.Set(ctx, key, val, c.cfg.RedisConfig.TTL)

					return nil
				})

				if err != nil {
					return fmt.Errorf("error setting merchant : %v", err)
				}

				return nil
			}

			return fmt.Errorf("error getting merchant : %v", err)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, oldVal, c.cfg.RedisConfig.TTL)
			return nil
		})

		c.log.Info("new message in redis")

		if err != nil {
			return err
		}

		return nil
	}, key)

	if err != nil {
		return fmt.Errorf("error transaction merchant : %v", err)
	}

	return nil
}

func (c cache) DeleteMerchant(ctx context.Context, key, keyUserID string) error {
	err := c.client.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)
			pipe.Del(ctx, keyUserID)
			return nil
		})

		return err

	}, key, keyUserID)

	if err != nil {
		return fmt.Errorf("error transaction merchant : %v", err)
	}

	return nil
}
