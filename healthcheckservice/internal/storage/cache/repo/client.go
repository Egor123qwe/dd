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

func (c cache) DeleteClientRent(ctx context.Context, key, sessionID string) error {
	var rents []model.Client

	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	err := c.client.Watch(ctx, func(tx *redis.Tx) error {
		oldVal, err := tx.Get(ctx, key).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				return fmt.Errorf("error client has no rents: %v", err)
			}

			return fmt.Errorf("error getting client : %v", err)
		}

		err = json.Unmarshal([]byte(oldVal), &rents)
		if err != nil {
			return fmt.Errorf("error unmarshalling client : %v", err)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {

			for k := range rents {
				if rents[k].SessionID == sessionID {
					rents = append(rents[:k], rents[k+1:]...)
					break
				}
			}

			if len(rents) == 0 {
				pipe.Del(ctx, key)
				return nil
			}

			val, err := json.Marshal(&rents)
			if err != nil {
				return fmt.Errorf("error marshalling client : %v", err)
			}

			pipe.Set(ctx, key, val, c.cfg.RedisConfig.TTL)
			return nil
		})

		c.log.Info("new message in redis")

		return err
	}, key)

	if err != nil {
		return fmt.Errorf("error transaction client : %v", err)
	}

	return nil
}

func (c cache) SetOrUpdateClient(ctx context.Context, key string, msg message.ClientRent) error {
	clientRent := model.Client{
		RequestID: msg.Content.RequestID,
		SessionID: msg.Content.SessionID,
		CreatedAt: msg.Content.CreatedAt,
		Status:    msg.Content.Status,
	}

	rents := make([]model.Client, 0)

	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	err := c.client.Watch(ctx, func(tx *redis.Tx) error {
		oldVal, err := tx.Get(ctx, key).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
					rents = append(rents, clientRent)

					val, err := json.Marshal(rents)
					if err != nil {
						return fmt.Errorf("error marshalling rents : %v", err)
					}

					pipe.Set(ctx, key, val, c.cfg.RedisConfig.TTL)

					return nil
				})

				if err != nil {
					return fmt.Errorf("error setting client : %v", err)
				}

				return nil
			}

			return fmt.Errorf("error getting client : %v", err)
		}

		err = json.Unmarshal([]byte(oldVal), &rents)
		if err != nil {
			return fmt.Errorf("error unmarshalling client : %v", err)
		}

		switch msg.Type {
			
		case message.EventTypeSessionUpdated:
			exist := false

			for k := range rents {
				if rents[k].SessionID == msg.Content.SessionID {
					rents[k].Status = msg.Content.Status
					exist = true
					break
				}
			}

			if !exist {
				rents = append(rents, clientRent)
			}

		case message.EventTypeClientStartRent:
			for k := range rents {
				if rents[k].SessionID == msg.Content.SessionID {
					rents[k].Settings = msg.Content.Settings
					rents[k].Status = string(message.RentStatusRunnuing)
					break
				}
			}

		default:
			return fmt.Errorf("unsupported message type : %s", msg.Type)
		}

		clientValue, err := json.Marshal(&rents)
		if err != nil {
			return fmt.Errorf("error marshalling client : %v", err)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, clientValue, c.cfg.RedisConfig.TTL)

			return nil
		})

		c.log.Info("new message in redis")

		if err != nil {
			return err
		}

		return nil

	}, key)

	if err != nil {
		return fmt.Errorf("error transaction client : %v", err)
	}

	return nil
}

func (c cache) GetClientRentBySession(ctx context.Context, key, sessionID string) (model.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return model.Client{}, fmt.Errorf("error getting client : %w", err)
	}

	var rents []model.Client
	err = json.Unmarshal([]byte(val), &rents)
	if err != nil {
		return model.Client{}, fmt.Errorf("error unmarshalling client : %w", err)
	}

	for _, rent := range rents {
		if rent.SessionID == sessionID {
			return rent, nil
		}
	}

	return model.Client{}, ErrNoRent
}
