package cache

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type UserCache struct {
	Adrr        string
	Password    string
	DB          int
	ExpiredTime time.Duration
	Client      *redis.Client
}

func NewCache(adrr string, password string, db int, expireTime time.Duration) *UserCache {
	cache := &UserCache{
		Adrr:        adrr,
		Password:    password,
		DB:          db,
		ExpiredTime: expireTime,
		Client: redis.NewClient(&redis.Options{
			Addr:     adrr,
			Password: password,
			DB:       db,
		}),
	}

	return cache
}

func (cache *UserCache) getClient() *redis.Client {
	return cache.Client
}

func (cache *UserCache) Set(key string, value *UserInfo) error {
	client := cache.getClient()

	json, err := json.Marshal(value)
	if err != nil {
		return err
	}
	client.Set(key, json, cache.ExpiredTime)
	return nil
}

func (cache *UserCache) Get(key string) (*UserInfo, error) {
	client := cache.getClient()
	val, err := client.Get(key).Result()

	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{}
	err = json.Unmarshal([]byte(val), &userInfo)

	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (cache *UserCache) Del(key string) {
	client := cache.getClient()
	client.Del(key)
}
