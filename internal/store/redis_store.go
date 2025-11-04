package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyDelTasks = "tgref:delete_tasks" // ZSET(score=execAtUnix, member="chatID:msgID")

type Task struct {
	ChatID    int64
	MsgID     int
	ExecuteAt time.Time
}

type RedisStore struct {
	rdb *redis.Client
}

func NewRedis(addr, pass string, db int) *RedisStore {
	return &RedisStore{
		rdb: redis.NewClient(&redis.Options{Addr: addr, Password: pass, DB: db}),
	}
}

// ScheduleDeletion сохраняет задачу удаления
func (s *RedisStore) ScheduleDeletion(ctx context.Context, chatID int64, msgID int, execAt time.Time) error {
	member := fmt.Sprintf("%d:%d", chatID, msgID)
	score := float64(execAt.Unix())
	return s.rdb.ZAdd(ctx, keyDelTasks, redis.Z{Score: score, Member: member}).Err()
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
