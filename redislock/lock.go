package redislock

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"sync"
	"time"
)

type Locker interface {
	Lock(string) error
	Unlock(string) error
}

type defaultLock struct {
	client redis.UniversalClient

	// lock timeout
	expiration time.Duration
	// allow network delay
	drift time.Duration

	val int32
	l   *sync.RWMutex
}

var (
	lockScript = `
	if redis.call("EXISTS", KEYS[1]) == 1 then
		return 0
	end
	return redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
	`

	unlockScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`

	ErrTaken    = errors.New("lock has been taken")
	ErrReleased = errors.New("lock has been released")
)

func (dl *defaultLock) formatKey(key string) string {
	return fmt.Sprintf("%s-lock", key)
}

func (dl *defaultLock) Lock(key string) error {
	skey := dl.formatKey(key)
	start := time.Now()
	val := rand.Int31()
	ctx, _ := context.WithTimeout(context.Background(), 2*dl.drift)

	res, err := dl.client.Eval(ctx, lockScript, []string{skey}, val, dl.expiration.Milliseconds()).Result()
	if err != nil {
		return err
	}
	success := res == "OK"

	if success && dl.expiration+dl.drift-time.Since(start) > 0 {
		dl.l.Lock()
		dl.val = val
		dl.l.Unlock()
		return nil
	}

	return ErrTaken
}

func (dl *defaultLock) Unlock(key string) error {
	skey := dl.formatKey(key)
	ctx, _ := context.WithTimeout(context.Background(), 2*dl.drift)

	dl.l.RLock()
	defer dl.l.RUnlock()
	res, err := dl.client.Eval(ctx, unlockScript, []string{skey}, dl.val).Result()
	if err != nil {
		return err
	}
	if res != int64(0) {
		return nil
	}

	return ErrReleased
}

func NewLocker(cli redis.UniversalClient) Locker {
	return &defaultLock{
		client:     cli,
		drift:      2 * time.Second,
		expiration: 10 * time.Second,
		l:          &sync.RWMutex{},
	}
}
