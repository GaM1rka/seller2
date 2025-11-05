package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	rdb *redis.Client
}

const (
	keyDelTasks   = "tgref:delete_tasks" // ZSET(score=execAtUnix, member="chatID:msgID")
	keyOfferTasks = "tgref:offer_tasks"  // ZSET(score=execAtUnix, member="chatID:msgID")
)

type Task struct {
	ChatID    int64
	MsgID     int
	ExecuteAt time.Time
}

func (s *RedisStore) ScheduleDeletion(ctx context.Context, chatID int64, msgID int, execAt time.Time) error {
	member := fmt.Sprintf("%d:%d", chatID, msgID)
	return s.rdb.ZAdd(ctx, keyDelTasks, redis.Z{Score: float64(execAt.Unix()), Member: member}).Err()
}

func (s *RedisStore) ScheduleOffer(ctx context.Context, chatID int64, msgID int, execAt time.Time) error {
	member := fmt.Sprintf("%d:%d", chatID, msgID)
	return s.rdb.ZAdd(ctx, keyOfferTasks, redis.Z{Score: float64(execAt.Unix()), Member: member}).Err()
}

func (s *RedisStore) FetchDueDeletions(ctx context.Context, now time.Time, limit int64) ([]Task, error) {
	return s.fetchAndPop(ctx, keyDelTasks, now, limit)
}
func (s *RedisStore) FetchDueOffers(ctx context.Context, now time.Time, limit int64) ([]Task, error) {
	return s.fetchAndPop(ctx, keyOfferTasks, now, limit)
}

func (s *RedisStore) fetchAndPop(ctx context.Context, key string, now time.Time, limit int64) ([]Task, error) {
	max := fmt.Sprintf("%d", now.Unix())
	ids, err := s.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: "-inf", Max: max, Offset: 0, Count: limit}).Result()
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	pipe := s.rdb.TxPipeline()
	for _, m := range ids {
		pipe.ZRem(ctx, key, m)
	}
	_, _ = pipe.Exec(ctx)

	tasks := make([]Task, 0, len(ids))
	for _, m := range ids {
		parts := strings.Split(m, ":")
		if len(parts) != 2 {
			continue
		}
		chat, _ := strconv.ParseInt(parts[0], 10, 64)
		msg, _ := strconv.Atoi(parts[1])
		tasks = append(tasks, Task{ChatID: chat, MsgID: msg})
	}
	return tasks, nil
}

func NewRedis(addr, pass string, db int) *RedisStore {
	return &RedisStore{
		rdb: redis.NewClient(&redis.Options{Addr: addr, Password: pass, DB: db}),
	}
}

// FetchDue забирает просроченные задачи (score <= now) до limit штук и удаляет их из ZSET
func (s *RedisStore) FetchDue(ctx context.Context, now time.Time, limit int64) ([]Task, error) {
	max := fmt.Sprintf("%d", now.Unix())
	ids, err := s.rdb.ZRangeByScore(ctx, keyDelTasks, &redis.ZRangeBy{Min: "-inf", Max: max, Offset: 0, Count: limit}).Result()
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	// удаляем взятые (идемпотентно)
	pipe := s.rdb.TxPipeline()
	for _, m := range ids {
		pipe.ZRem(ctx, keyDelTasks, m)
	}
	_, _ = pipe.Exec(ctx)

	tasks := make([]Task, 0, len(ids))
	for _, m := range ids {
		parts := strings.Split(m, ":")
		if len(parts) != 2 {
			continue
		}
		chat, _ := strconv.ParseInt(parts[0], 10, 64)
		msg, _ := strconv.Atoi(parts[1])
		tasks = append(tasks, Task{ChatID: chat, MsgID: msg})
	}
	return tasks, nil
}

func (s *RedisStore) Ping(ctx context.Context) (string, error) {
	return s.rdb.Ping(ctx).Result()
}
