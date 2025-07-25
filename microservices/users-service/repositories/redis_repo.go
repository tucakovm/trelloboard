package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"
	"users_module/customLogger"

	"github.com/redis/go-redis/v9"
	"users_module/models"
)

type RedisRepo struct {
	Cli *redis.Client
}

func NewRedisRepo() *RedisRepo {
	logger := customLogger.GetLogger()
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"), // Replace with your Redis server address
		Password: os.Getenv("PASSWORD"),      // No password by default
		DB:       0,                          // Default DB
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		logger.Error(nil, "Error connecting to Redis: "+err.Error())
		log.Fatalf("Error connecting to Redis: %v", err)

	}

	logger.Info(nil, "Connected to Redis successfully")
	return &RedisRepo{Cli: client}
}

func (r *RedisRepo) SaveUnverifiedUser(user models.User, expiration time.Duration) error {
	logger := customLogger.GetLogger()
	ctx := context.Background()

	data, err := json.Marshal(user)
	if err != nil {
		logger.Error(nil, "JSON marshal error for unverified user: "+err.Error())
		return err
	}
	logger.Info(map[string]interface{}{"username": user.Username}, "Unverified user marshalled to JSON")

	key := "unverified_user:" + user.Username

	err = r.Cli.Set(ctx, key, data, expiration).Err()
	if err != nil {
		logger.Error(map[string]interface{}{"username": user.Username}, "Failed to save unverified user in Redis: "+err.Error())
		return err
	}

	return nil
}

func (r *RedisRepo) GetUnverifiedUser(username string) (*models.User, error) {
	logger := customLogger.GetLogger()
	ctx := context.Background()

	key := "unverified_user:" + username
	data, err := r.Cli.Get(ctx, key).Result()
	if err == redis.Nil {
		logger.Warn(map[string]interface{}{"username": username}, "Unverified user not found in Redis")
		log.Println("Unverified user not found")
		return nil, errors.New("user not found")
	} else if err != nil {
		logger.Error(map[string]interface{}{"username": username}, "Failed to fetch unverified user from Redis: "+err.Error())
		return nil, err
	}

	// Deserialize JSON to user struct
	var user models.User
	log.Println("Fetched unverified user", user.Username)
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		logger.Error(map[string]interface{}{"username": username}, "Error unmarshalling unverified user: "+err.Error())
		log.Println("Error unmarshalling unverified user")
		return nil, err
	}

	logger.Info(map[string]interface{}{"username": user.Username}, "Fetched unverified user from Redis successfully")
	return &user, nil
}

//func (r *RedisRepo) DeleteUnverifiedUser(username string) error {
//	ctx := context.Background()
//
//	key := "unverified_user:" + username
//	err := r.Cli.Del(ctx, key).Err()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (r *RedisRepo) SaveRecoveryCode(username string, code string, expiration time.Duration) error {
	logger := customLogger.GetLogger()
	ctx := context.Background()
	key := "recovery_code:" + username
	err := r.Cli.Set(ctx, key, code, expiration).Err()
	if err != nil {
		logger.Error(map[string]interface{}{"username": username}, "Failed to save recovery code: "+err.Error())
		return err
	}

	logger.Info(map[string]interface{}{"username": username}, "Saved recovery code in Redis")
	return nil
}

func (r *RedisRepo) GetRecoveryCode(username string) (string, error) {
	logger := customLogger.GetLogger()
	ctx := context.Background()
	key := "recovery_code:" + username
	log.Println("fetched recovery code for:", username)
	data, err := r.Cli.Get(ctx, key).Result()
	if err == redis.Nil {
		logger.Warn(map[string]interface{}{"username": username}, "Recovery code not found in Redis")
		log.Println("Recovery code not found")
	}
	if err != nil {
		logger.Error(map[string]interface{}{"username": username}, "Failed to get recovery code: "+err.Error())
		return "", err
	}

	logger.Info(map[string]interface{}{"username": username}, "Fetched recovery code from Redis")
	log.Println("recovery code:", data)
	return data, nil
}

func (r *RedisRepo) DeleteRecoveryCode(username string) error {
	logger := customLogger.GetLogger()
	ctx := context.Background()
	key := "recovery_code:" + username
	err := r.Cli.Del(ctx, key).Err()
	if err != nil {
		logger.Error(map[string]interface{}{"username": username}, "Error deleting recovery code: "+err.Error())
		log.Println("Error deleting recovery code")
		return err
	}

	logger.Info(map[string]interface{}{"username": username}, "Deleted recovery code from Redis")
	return nil
}
