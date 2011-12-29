package store

import (
	"os"
	"github.com/simonz05/godis"
)

type RedisStore struct {
	client *godis.Client
}

func NewRedisStore(addr string) *RedisStore {
	return &RedisStore{client: godis.New(addr, 0, "")}
}

func (s *RedisStore) Set(key string, val string, exp int64) os.Error {
	return s.client.Setex(key, exp, val)
}

func (s *RedisStore) Get(key string) (string, os.Error) {
	elem, err := s.client.Get(key)
	return elem.String(), err
}

func (s *RedisStore) Delete(keys ...string) {
	s.client.Del(keys...)
}