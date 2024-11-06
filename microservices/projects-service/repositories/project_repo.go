package repositories

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"projects_module/domain"
	"time"
)

type ProjectRepo struct {
	cli *mongo.Client
}

func (pr *ProjectRepo) Disconnect(ctx context.Context) error {
	err := pr.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (pr *ProjectRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := pr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Println(err)
	}

	// Print available databases
	databases, err := pr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Println(err)
	}
	fmt.Println(databases)
}

func (pr *ProjectRepo) getCollection() *mongo.Collection {
	patientDatabase := pr.cli.Database("mongoDemo")
	patientsCollection := patientDatabase.Collection("projects")
	return patientsCollection
}

func New(ctx context.Context, logger *log.Logger) (*ProjectRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &ProjectRepo{
		cli: client,
	}, nil
}

func (pr *ProjectRepo) Create(project *domain.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	patientsCollection := pr.getCollection()

	result, err := patientsCollection.InsertOne(ctx, &project)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (pr *ProjectRepo) GetAll() (domain.Projects, error) {
	// Initialise context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	patientsCollection := pr.getCollection()

	var project domain.Projects
	patientsCursor, err := patientsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if err = patientsCursor.All(ctx, &project); err != nil {
		log.Println(err)
		return nil, err
	}
	return project, nil
}

func (pr *ProjectRepo) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	patientsCollection := pr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := patientsCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}
