package repositories

import (
	"context"
	"errors"
	"fmt"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"time"
	"users_module/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct {
	Cli    *mongo.Client
	Tracer trace.Tracer
}

func NewUserRepo(ctx context.Context, tracer trace.Tracer) (*UserRepo, error) {
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

	return &UserRepo{Cli: client,
		Tracer: tracer}, nil
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

	id, err := primitive.ObjectIDFromHex("67386650a0d21b3a8f823722")
	if err != nil {
		fmt.Println("Error converting to ObjectID:", err)
		return nil
	}
	bobId, err := primitive.ObjectIDFromHex("6749ac3cfc079b8c923bb9d5")
	if err != nil {
		fmt.Println("Error converting to ObjectID:", err)
		return nil
	}

	users := []interface{}{
		models.User{
			Id:        id,
			FirstName: "Alice",
			LastName:  "Johnson",
			Username:  "alicej",
			Email:     "alice.johnson@example.com",
			Password:  "$2a$12$sH1miPk3Yk1umoZKoJGnDOGofZZzher2JFa1AUceFKTlx6Glcd64O",
			IsActive:  true,
			Code:      "A123",
			Role:      "Manager",
		},
		models.User{
			Id:        bobId,
			FirstName: "Bob",
			LastName:  "Smith",
			Username:  "bobsmith",
			Email:     "bob.smith@example.com",
			Password:  "$2a$12$sH1miPk3Yk1umoZKoJGnDOGofZZzher2JFa1AUceFKTlx6Glcd64O",
			IsActive:  true,
			Code:      "B456",
			Role:      "User",
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

func (tr *UserRepo) getCollection(ctx context.Context) *mongo.Collection {
	ctx, span := tr.Tracer.Start(ctx, "r.getCollection")
	defer span.End()
	if tr.Cli == nil {
		log.Println("Mongo client is nil!")
		return nil
	}

	if err := tr.Cli.Ping(context.Background(), nil); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error pinging MongoDB, connection lost:", err)
		return nil
	}

	return tr.Cli.Database("mongoDemo").Collection("users")
}

func (tr *UserRepo) SaveUser(user models.User, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.saveUser")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		return errors.New("failed to retrieve collection")
	}

	// Setting a new ObjectID for the user
	user.Id = primitive.NewObjectID()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error saving user:", err)
		return err
	}

	log.Println("User saved successfully:", user)
	return nil
}
func (tr *UserRepo) GetUserByUsername(username string, ctx context.Context) (*models.User, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.getUserByUsername")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		return nil, errors.New("failed to retrieve collection")
	}

	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("User not found with username:", username)
		return nil, err
	}

	return &user, nil
}
func (tr *UserRepo) GetUserByEmail(email string, ctx context.Context) (*models.User, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.getUserByEmail")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		return nil, errors.New("failed to retrieve collection")
	}

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("User not found with email:", email)
		return nil, ErrUserNotFound
	}

	return &user, nil
}

func (tr *UserRepo) ActivateUser(username string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.activateUser")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"is_active": true}}
	log.Println("active user inside user_repo")
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error activating user:", err)
		return err
	}

	log.Printf("User with email %s activated successfully", username)
	return nil
}

func (tr *UserRepo) GetAll(ctx context.Context) ([]models.User, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.getAll")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		return nil, fmt.Errorf("failed to retrieve collection")
	}

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error iterating over cursor:", err)
		return nil, err
	}

	return tasks, nil
}

func (tr *UserRepo) Delete(username string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.delete")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	filter := bson.M{"username": username}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error deleting user:", err)
		return err
	}

	if result.DeletedCount == 0 {
		log.Println("No user found with username:", username)
		return fmt.Errorf("no user found with the provided username")
	}

	log.Printf("User with username %s deleted successfully", username)
	return nil
}

func (tr *UserRepo) DeleteById(id string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.deleteById")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	filter := bson.M{"username": id}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error deleting user:", err)
		return err
	}

	if result.DeletedCount == 0 {
		log.Println("No user found with id:", id)
		return fmt.Errorf("no user found with the provided id")
	}

	log.Printf("User deleted successfully", id)
	return nil
}

func (tr *UserRepo) UpdatePassword(username, hashedPassword string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.updatePass")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		log.Println("Failed to retrieve collection")
		return fmt.Errorf("failed to retrieve collection")
	}

	_, err := collection.UpdateOne(ctx, bson.M{"username": username}, bson.M{"$set": bson.M{"password": hashedPassword}})
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Failed to update password for user:", username)
		return err
	}

	return nil
}

func (pr *UserRepo) Disconnect(ctx context.Context) error {
	ctx, span := pr.Tracer.Start(ctx, "r.disconnect")
	defer span.End()
	err := pr.Cli.Disconnect(ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return err
	}
	return nil
}
