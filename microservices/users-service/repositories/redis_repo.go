package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"users_module/models"
)

type RedisRepo struct {
	Cli *redis.Client
}

func NewRedisRepo() *RedisRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Replace with your Redis server address
		Password: "",               // No password by default
		DB:       0,                // Default DB
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}

	return &RedisRepo{Cli: client}
}

func (r *RedisRepo) SaveUnverifiedUser(user models.User, expiration time.Duration) error {
	ctx := context.Background()

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	log.Println("proso json marshal")

	key := "unverified_user:" + user.Username

	err = r.Cli.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisRepo) GetUnverifiedUser(username string) (*models.User, error) {
	ctx := context.Background()

	key := "unverified_user:" + username
	data, err := r.Cli.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Println("Unverified user not found")
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}

	// Deserialize JSON to user struct
	var user models.User
	log.Println("Fetched unverified user", user.Username)
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		log.Println("Error unmarshalling unverified user")
		return nil, err
	}

	return &user, nil
}

func (r *RedisRepo) DeleteUnverifiedUser(email string) error {
	ctx := context.Background()

	key := "unverified_user:" + email
	err := r.Cli.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}
