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
	"users_module/customLogger"
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
	logger := customLogger.GetLogger()
	dburi := os.Getenv("MONGO_DB_URI")
	if dburi == "" {
		logger.Error(nil, "MONGO_DB_URI is not set")
		return nil, fmt.Errorf("MONGO_DB_URI is not set")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		logger.Error(nil, "Failed to connect to MongoDB: "+err.Error())
		log.Printf("Failed to connect to MongoDB: %v", err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		logger.Error(nil, "MongoDB ping failed: "+err.Error())
		log.Printf("MongoDB ping failed: %v", err)
		return nil, err
	}

	logger.Info(nil, "Connected to MongoDB successfully")
	log.Println("Connected to MongoDB successfully")

	if err := insertInitialUsers(client); err != nil {
		logger.Error(nil, "Failed to insert initial users: "+err.Error())
		log.Printf("Failed to insert initial tasks: %v", err)
	}

	return &UserRepo{Cli: client,
		Tracer: tracer}, nil
}

func insertInitialUsers(client *mongo.Client) error {
	logger := customLogger.GetLogger()
	collection := client.Database("mongoDemo").Collection("users")
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		logger.Error(nil, "Error checking user count: "+err.Error())
		log.Println("Error checking user count:", err)
		return err
	}

	if count > 0 {
		logger.Info(nil, "Users already exist in the database")
		log.Println("Users already exist in the database")
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
	daveId, err := primitive.ObjectIDFromHex("676b0c5dee23d6f7b4ff6789")
	if err != nil {
		fmt.Println("Error converting to ObjectID for Dave:", err)
		return nil
	}
	janeId, err := primitive.ObjectIDFromHex("6771ac3cfc079b8c923bb9d5")
	if err != nil {
		fmt.Println("Error converting to ObjectID for Jane:", err)
		return nil
	}
	lilyId, err := primitive.ObjectIDFromHex("6791ac3cfc079b8c923bb9d5")
	if err != nil {
		fmt.Println("Error converting to ObjectID for Lily:", err)
		return nil
	}
	markId, err := primitive.ObjectIDFromHex("6781ac3cfc079b8c923bb9d5")
	if err != nil {
		fmt.Println("Error converting to ObjectID for Mark:", err)
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
		models.User{
			Id:        daveId,
			FirstName: "Dave",
			LastName:  "White",
			Username:  "davew",
			Email:     "dave.white@example.com",
			Password:  "$2a$12$sH1miPk3Yk1umoZKoJGnDOGofZZzher2JFa1AUceFKTlx6Glcd64O",
			IsActive:  true,
			Code:      "D012",
			Role:      "User",
		},
		models.User{
			Id:        janeId,
			FirstName: "Jane",
			LastName:  "Doe",
			Username:  "janedoe",
			Email:     "jane.doe@example.com",
			Password:  "$2a$12$sH1miPk3Yk1umoZKoJGnDOGofZZzher2JFa1AUceFKTlx6Glcd64O",
			IsActive:  true,
			Code:      "J999",
			Role:      "Manager",
		},
		models.User{
			Id:        lilyId,
			FirstName: "Lily",
			LastName:  "Evans",
			Username:  "lilyevans",
			Email:     "lily.evans@example.com",
			Password:  "$2a$12$sH1miPk3Yk1umoZKoJGnDOGofZZzher2JFa1AUceFKTlx6Glcd64O",
			IsActive:  true,
			Code:      "L555",
			Role:      "User",
		},
		models.User{
			Id:        markId,
			FirstName: "Mark",
			LastName:  "Taylor",
			Username:  "marktaylor",
			Email:     "mark.taylor@example.com",
			Password:  "$2a$12$sH1miPk3Yk1umoZKoJGnDOGofZZzher2JFa1AUceFKTlx6Glcd64O",
			IsActive:  true,
			Code:      "M123",
			Role:      "User",
		},
	}

	// Insert initial users
	_, err = collection.InsertMany(context.Background(), users)
	if err != nil {
		logger.Error(nil, "Error inserting initial users: "+err.Error())
		log.Println("Error inserting initial users:", err)
		return err
	}

	logger.Info(nil, "Initial users inserted successfully")
	log.Println("Initial users inserted successfully")
	return nil
}

func (tr *UserRepo) getCollection(ctx context.Context) *mongo.Collection {
	logger := customLogger.GetLogger()
	ctx, span := tr.Tracer.Start(ctx, "r.getCollection")
	defer span.End()
	if tr.Cli == nil {
		logger.Error(nil, "Mongo client is nil!")
		log.Println("Mongo client is nil!")
		return nil
	}

	if err := tr.Cli.Ping(context.Background(), nil); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		logger.Error(nil, "Mongo ping failed: "+err.Error())
		log.Println("Error pinging MongoDB, connection lost:", err)
		return nil
	}

	logger.Info(nil, "Mongo collection retrieved")
	return tr.Cli.Database("mongoDemo").Collection("users")
}

func (tr *UserRepo) SaveUser(user models.User, ctx context.Context) error {
	logger := customLogger.GetLogger()
	ctx, span := tr.Tracer.Start(ctx, "r.saveUser")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		logger.Error(nil, "Collection not found")
		return errors.New("failed to retrieve collection")
	}

	// Setting a new ObjectID for the user
	user.Id = primitive.NewObjectID()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		logger.Error(map[string]interface{}{"username": user.Username}, "Error saving user: "+err.Error())
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error saving user:", err)
		return err
	}

	logger.Info(map[string]interface{}{"username": user.Username}, "User saved successfully")
	log.Println("User saved successfully:", user)
	return nil
}
func (tr *UserRepo) GetUserByUsername(username string, ctx context.Context) (*models.User, error) {
	logger := customLogger.GetLogger()
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
		logger.Warn(map[string]interface{}{"username": username}, "User not found")
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("User not found with username:", username)
		return nil, err
	}

	logger.Info(map[string]interface{}{"username": username}, "User fetched")
	return &user, nil
}
func (tr *UserRepo) GetUserByEmail(email string, ctx context.Context) (*models.User, error) {
	logger := customLogger.GetLogger()
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
		logger.Warn(map[string]interface{}{"email": email}, "User not found by email")
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("User not found with email:", email)
		return nil, ErrUserNotFound
	}

	logger.Info(map[string]interface{}{"email": email}, "User fetched by email")
	return &user, nil
}

func (tr *UserRepo) ActivateUser(username string, ctx context.Context) error {
	logger := customLogger.GetLogger()
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
		logger.Error(map[string]interface{}{"username": username}, "Error activating user: "+err.Error())
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error activating user:", err)
		return err
	}

	logger.Info(map[string]interface{}{"username": username}, "User activated")
	log.Printf("User with email %s activated successfully", username)
	return nil
}

func (tr *UserRepo) GetAll(ctx context.Context) ([]models.User, error) {
	logger := customLogger.GetLogger()
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
		logger.Error(nil, "Error finding users: "+err.Error())
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error finding tasks:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []models.User
	for cursor.Next(ctx) {
		var task models.User
		if err := cursor.Decode(&task); err != nil {
			logger.Error(nil, "Error decoding user: "+err.Error())
			log.Println("Error decoding task:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		logger.Error(nil, "Cursor error: "+err.Error())
		log.Println("Error iterating over cursor:", err)
		return nil, err
	}

	logger.Info(map[string]interface{}{"user_count": len(tasks)}, "Fetched all users")
	return tasks, nil
}

func (tr *UserRepo) Delete(username string, ctx context.Context) error {
	logger := customLogger.GetLogger()
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
		logger.Error(map[string]interface{}{"username": username}, "Error deleting user: "+err.Error())
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error deleting user:", err)
		return err
	}

	if result.DeletedCount == 0 {
		logger.Warn(map[string]interface{}{"username": username}, "No user found to delete")
		log.Println("No user found with username:", username)
		return fmt.Errorf("no user found with the provided username")
	}

	logger.Info(map[string]interface{}{"username": username}, "User deleted successfully")
	log.Printf("User with username %s deleted successfully", username)
	return nil
}

func (tr *UserRepo) DeleteById(id string, ctx context.Context) error {
	logger := customLogger.GetLogger()
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
		logger.Error(map[string]interface{}{"id": id}, "Error deleting user by ID: "+err.Error())
		log.Println("Error deleting user:", err)
		return err
	}

	if result.DeletedCount == 0 {
		log.Println("No user found with id:", id)
		logger.Warn(map[string]interface{}{"id": id}, "No user found with given ID")
		return fmt.Errorf("no user found with the provided id")
	}

	logger.Info(map[string]interface{}{"id": id}, "User deleted by ID successfully")
	log.Printf("User deleted successfully", id)
	return nil
}

func (tr *UserRepo) UpdatePassword(username, hashedPassword string, ctx context.Context) error {
	logger := customLogger.GetLogger()
	ctx, span := tr.Tracer.Start(ctx, "r.updatePass")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	collection := tr.getCollection(ctx)
	if collection == nil {
		logger.Error(nil, "Collection not found")
		log.Println("Failed to retrieve collection")
		return fmt.Errorf("failed to retrieve collection")
	}

	_, err := collection.UpdateOne(ctx, bson.M{"username": username}, bson.M{"$set": bson.M{"password": hashedPassword}})
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		logger.Error(map[string]interface{}{"username": username}, "Failed to update password")
		log.Println("Failed to update password for user:", username)
		return err
	}

	logger.Info(map[string]interface{}{"username": username}, "Password updated successfully")
	return nil
}

func (pr *UserRepo) Disconnect(ctx context.Context) error {
	logger := customLogger.GetLogger()
	ctx, span := pr.Tracer.Start(ctx, "r.disconnect")
	defer span.End()
	err := pr.Cli.Disconnect(ctx)
	if err != nil {
		logger.Error(nil, "Mongo disconnect failed: "+err.Error())
		span.SetStatus(otelCodes.Error, err.Error())
		return err
	}

	logger.Info(nil, "Mongo disconnected successfully")
	return nil
}
