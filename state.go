package main

import (
	"fmt"
	"log"

	"gopkg.in/redis.v5"
)

type State interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Clear(namespace string) error
}

type RedisState struct {
	client *redis.Client
}

func NewRedisState(addr string) (*RedisState, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &RedisState{
		client: redisClient,
	}, nil
}

func (r *RedisState) Set(key, value string) error {
	res := r.client.Set(key, value, -1)
	return res.Err()
}

func (r *RedisState) Get(key string) (string, error) {
	res := r.client.Get(key)
	return res.Val(), res.Err()
}

func (r *RedisState) Clear(namespace string) error {
	keysCmd := r.client.Keys(fmt.Sprintf("k8s-deployer/%s/*", namespace))
	keys, err := keysCmd.Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		log.Println("Removing:", key)
		r.client.Del(key)
	}

	return nil
}
