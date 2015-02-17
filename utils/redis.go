package utils

import (
	"bytes"
	"errors"
	"github.com/garyburd/redigo/redis"
	"time"

	"github.com/techjanitor/pram-post/config"
)

// RedisStore holds a handle to the Redis pool
type RedisStore struct {
	pool *redis.Pool
}

var (
	RedisCache   RedisStore
	ErrCacheMiss = errors.New("cache: key not found.")
	buffer       bytes.Buffer
)

// NewRedisCache creates a new pool
func NewRedisCache() {
	RedisCache.pool = &redis.Pool{
		MaxIdle:     config.Settings.Redis.MaxIdle,
		MaxActive:   config.Settings.Redis.MaxConnections,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(config.Settings.Redis.Protocol, config.Settings.Redis.Address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}

	return
}

// Get will retrieve a key
func (c *RedisStore) Get(key string) (result []byte, err error) {
	conn := c.pool.Get()
	defer conn.Close()

	raw, err := conn.Do("GET", key)
	if raw == nil {
		return nil, ErrCacheMiss
	}
	result, err = redis.Bytes(raw, err)
	if err != nil {
		return
	}

	return
}

// HGet will retrieve a hash
func (c *RedisStore) HGet(key string, value string) (result []byte, err error) {
	conn := c.pool.Get()
	defer conn.Close()

	raw, err := conn.Do("HGET", key, value)
	if raw == nil {
		return nil, ErrCacheMiss
	}
	result, err = redis.Bytes(raw, err)
	if err != nil {
		return
	}

	return
}

// Set will set a single record
func (c *RedisStore) Set(key string, result []byte) (err error) {
	conn := c.pool.Get()
	defer conn.Close()

	_, err = conn.Do("SET", key, result)

	return
}

// Set will set a single record
func (c *RedisStore) SetEx(key string, timeout uint, result []byte) (err error) {
	conn := c.pool.Get()
	defer conn.Close()

	_, err = conn.Do("SETEX", key, timeout, result)

	return
}

// HMSet will set a hash
func (c *RedisStore) HMSet(key string, value string, result []byte) (err error) {
	conn := c.pool.Get()
	defer conn.Close()

	_, err = conn.Do("HMSET", key, value, result)

	return
}

// Delete will delete a key
func (c *RedisStore) Delete(key ...interface{}) (err error) {
	conn := c.pool.Get()
	defer conn.Close()

	_, err = conn.Do("DEL", key...)

	return
}

// Flush will call flushall and delete all keys
func (c *RedisStore) Flush() (err error) {
	conn := c.pool.Get()
	defer conn.Close()

	_, err = conn.Do("FLUSHALL")

	return
}
