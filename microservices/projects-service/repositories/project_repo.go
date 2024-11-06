package repositories

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
	projectsDatabase := pr.cli.Database("mongoDemo")
	patientsCollection := projectsDatabase.Collection("projects")
	return patientsCollection
}

func New(ctx context.Context, logger *log.Logger) (*ProjectRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	// Optionally, check if the connection is valid by pinging the database
	err = client.Ping(ctx, nil)
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
	projectsCollection := pr.getCollection()

	result, err := projectsCollection.InsertOne(ctx, &project)
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
	log.Println("test1")
	projectsCollection := pr.getCollection()

	var project domain.Projects
	log.Println("test2")
	patientsCursor, err := projectsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("test3")
	if err = patientsCursor.All(ctx, &project); err != nil {
		log.Println(err)
		return nil, err
	}
	return project, nil
}

func (pr *ProjectRepo) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	projectsCollection := pr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := projectsCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}
