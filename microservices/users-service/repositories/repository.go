package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"users_module/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct {
	Cli *mongo.Client
}

func NewUserRepo(ctx context.Context) (*UserRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")
	if dburi == "" {
		return nil, fmt.Errorf("MONGO_DB_URI is not set")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("MongoDB ping failed: %v", err)
		return nil, err
	}

	log.Println("Connected to MongoDB successfully")

	if err := insertInitialUsers(client); err != nil {
		log.Printf("Failed to insert initial tasks: %v", err)
	}

	return &UserRepo{Cli: client}, nil
}

func insertInitialUsers(client *mongo.Client) error {
	collection := client.Database("mongoDemo").Collection("users")
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println("Error checking task count:", err)
		return err
	}

	if count > 0 {
		log.Println("Tasks already exist in the database")
		return nil
	}

	users := []interface{}{
		models.User{
			Id:        primitive.NewObjectID(),
			FirstName: "Alice",
			LastName:  "Johnson",
			Username:  "alicej",
			Email:     "alice.johnson@example.com",
			Password:  "sifra123",
			IsActive:  true,
			Code:      "A123",
		},
		models.User{
			Id:        primitive.NewObjectID(),
			FirstName: "Bob",
			LastName:  "Smith",
			Username:  "bobsmith",
			Email:     "bob.smith@example.com",
			Password:  "sifra12345",
			IsActive:  false,
			Code:      "B456",
		},
	}

	// Insert initial tasks
	_, err = collection.InsertMany(context.Background(), users)
	if err != nil {
		log.Println("Error inserting initial tasks:", err)
		return err
	}

	log.Println("Initial tasks inserted successfully")
	return nil
}

func (tr *UserRepo) getCollection() *mongo.Collection {
	if tr.Cli == nil {
		log.Println("Mongo client is nil!")
		return nil
	}

	if err := tr.Cli.Ping(context.Background(), nil); err != nil {
		log.Println("Error pinging MongoDB, connection lost:", err)
		return nil
	}

	return tr.Cli.Database("mongoDemo").Collection("users")
}

func (tr *UserRepo) SaveUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return errors.New("failed to retrieve collection")
	}

	// Setting a new ObjectID for the user
	user.Id = primitive.NewObjectID()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Println("Error saving user:", err)
		return err
	}

	log.Println("User saved successfully:", user)
	return nil
}
func (tr *UserRepo) GetUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return nil, errors.New("failed to retrieve collection")
	}

	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		log.Println("User not found with username:", username)
		return nil, err
	}

	return &user, nil
}
func (tr *UserRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return nil, errors.New("failed to retrieve collection")
	}

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Println("User not found with email:", email)
		return nil, ErrUserNotFound
	}

	return &user, nil
}

func (tr *UserRepo) ActivateUser(email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"is_active": true}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Error activating user:", err)
		return err
	}

	log.Printf("User with email %s activated successfully", email)
	return nil
}
func (tr *UserRepo) GetAll() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return nil, fmt.Errorf("failed to retrieve collection")
	}

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Error finding tasks:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []models.User
	for cursor.Next(ctx) {
		var task models.User
		if err := cursor.Decode(&task); err != nil {
			log.Println("Error decoding task:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		log.Println("Error iterating over cursor:", err)
		return nil, err
	}

	return tasks, nil
}

func (tr *UserRepo) Delete(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	filter := bson.M{"id": id}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println("Error deleting user:", err)
		return err
	}

	if result.DeletedCount == 0 {
		log.Println("No user found with ID:", id)
		return fmt.Errorf("no user found with the provided ID")
	}

	log.Printf("User with ID %s deleted successfully", id)
	return nil
}

func (pr *UserRepo) Disconnect(ctx context.Context) error {
	err := pr.Cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}
