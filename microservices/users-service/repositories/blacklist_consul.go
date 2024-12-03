package repositories

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
)

type BlacklistConsul struct {
	Client *api.Client
}

func NewBlacklistConsul(address string) (*BlacklistConsul, error) {
	client, err := api.NewClient(&api.Config{
		Address: address,
	})
	if err != nil {
		log.Println("Error creating Consul client:", err)
		return nil, err
	}

	if err := initializeBlacklist(client); err != nil {
		log.Printf("Failed to initialize blacklist: %v", err)
		return nil, err
	}

	// Perform checks or setup as needed
	return &BlacklistConsul{Client: client}, nil
}

func initializeBlacklist(client *api.Client) error {
	kv := client.KV()

	passwords := []string{
		"Password123!", "123456Abc!", "Qwerty1234!", "Welcome2024!", "Admin2023!",
		"Sunshine2024!", "QwertyUIOP1", "Password1!", "Test1234!", "Iloveyou2024!",
		"12345ABCDE!", "Monkey2024!", "HelloWorld1", "Superman2023!", "1Qaz2Wsx!",
	}

	// Check if any keys exist under the "blacklist/" prefix
	pairs, _, err := kv.List("blacklist/", nil)
	if err != nil {
		log.Printf("Error checking blacklist keys: %v", err)
		return err
	}

	if len(pairs) > 0 {
		log.Println("Blacklist already initialized")
		return nil
	}

	// Add each password to the "blacklist/" prefix
	for i, password := range passwords {
		key := fmt.Sprintf("blacklist/password-%d", i+1)
		entry := &api.KVPair{
			Key:   key,
			Value: []byte(password),
		}

		_, err := kv.Put(entry, nil)
		if err != nil {
			log.Printf("Error adding key to blacklist: %v", err)
			return err
		}
		log.Printf("Added key: %s with value: %s to the blacklist", key, password)
	}

	log.Println("Blacklist initialized with 15 passwords")
	return nil
}

func (bc *BlacklistConsul) GetKey(key string) (string, error) {
	pair, _, err := bc.Client.KV().Get("blacklist/"+key, nil)
	if err != nil {
		log.Printf("Failed to retrieve key %s: %v", key, err)
		return "", err
	}

	if pair == nil {
		return "", fmt.Errorf("key %s not found in blacklist", key)
	}

	return string(pair.Value), nil
}

func checkBlacklist(client *api.Client, password string) error {
	kv := client.KV()

	// Fetch all blacklist keys and values
	pairs, _, err := kv.List("blacklist/", nil)
	if err != nil {
		log.Printf("Error retrieving blacklist keys: %v", err)
		return fmt.Errorf("unable to check blacklist: %w", err)
	}

	// Iterate over the list and check for matches
	for _, pair := range pairs {
		if string(pair.Value) == password {
			log.Printf("Password '%s' is blacklisted!", password)
			return fmt.Errorf("password '%s' is blacklisted", password)
		}
	}

	// No match found
	log.Printf("Password '%s' is not blacklisted", password)
	return nil
}

func (bc *BlacklistConsul) CheckPassword(password string) error {
	kv := bc.Client.KV()

	// Fetch all blacklist keys and values
	pairs, _, err := kv.List("blacklist/", nil)
	if err != nil {
		log.Printf("Error retrieving blacklist keys: %v", err)
		return fmt.Errorf("unable to check blacklist: %w", err)
	}

	// Iterate over the keys to find a match
	for _, pair := range pairs {
		if string(pair.Value) == password {
			log.Printf("Password '%s' is blacklisted!", password)
			return fmt.Errorf("password '%s' is blacklisted", password)
		}
	}

	// No match found
	log.Printf("Password '%s' is not blacklisted", password)
	return nil
}

func (bc *BlacklistConsul) GetAllKeys() (map[string]string, error) {
	pairs, _, err := bc.Client.KV().List("blacklist/", nil)
	if err != nil {
		log.Printf("Failed to list blacklist keys: %v", err)
		return nil, err
	}

	result := make(map[string]string)
	for _, pair := range pairs {
		result[pair.Key] = string(pair.Value)
	}

	return result, nil
}
