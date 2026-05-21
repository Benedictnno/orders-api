package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Benedictnno/orders-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(orderID uint64) string {
	return fmt.Sprintf("order:%d", orderID)
}

func (r *RedisRepo) Insert(ctx context.Context, order *model.Order) error {
	// Implementation for inserting order into Redis
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}
	key := orderIDKey(order.OrderID)
	txn := r.Client.TxPipeline()
	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order into Redis: %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add order key to orders set: %w", err)
	}
	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, orderID uint64) (*model.Order, error) {
	// Implementation for retrieving order from Redis by ID
	key := orderIDKey(orderID)
	data, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, ErrNotExist
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve order from Redis: %w", err)
	}
	var order model.Order
	err = json.Unmarshal([]byte(data), &order)
	if err != nil {
		return nil, fmt.Errorf("failed to decode order: %m", err)
	}
	return &order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)
	txn := r.Client.TxPipeline()
	err := txn.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete order from Redis: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove order key from orders set: %w", err)
	}
	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order *model.Order) error {
	// Implementation for updating order in Redis
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}
	key := orderIDKey(order.OrderID)
	err = r.Client.Set(ctx, key, string(data), 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update order in Redis: %w", err)
	}

	return r.Insert(ctx, order)
}

type FindAllPage struct {
	Size   uint
	Offset uint
}

type FindAllResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindAllResult, error) {
	// Implementation for retrieving all orders from Redis
	keys, cursor, err := r.Client.SScan(ctx, "orders", uint64(page.Offset), "*", int64(page.Size)).Result()
	if err != nil {
		return FindAllResult{}, fmt.Errorf("failed to retrieve order keys from Redis: %w", err)
	}
	if len(keys) == 0 {
		return FindAllResult{
			Orders: []model.Order{},
			Cursor: cursor,
		}, nil
	}
	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindAllResult{}, fmt.Errorf("failed to retrieve orders from Redis: %w", err)
	}
	orders := make([]model.Order, 0, len(xs))
	for i, x := range xs {
		x := x.(string)
		var order model.Order
		err = json.Unmarshal([]byte(x), &order)
		if err != nil {
			return FindAllResult{}, fmt.Errorf("failed to decode order with key %s: %w", keys[i], err)
		}
		orders[i] = order
	}
	return FindAllResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
