package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"projects_module/domain"
	proto "projects_module/proto/project"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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

func insertInitialProjects(client *mongo.Client) error {
	collection := client.Database("mongoDemo").Collection("projects")
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println("Error checking project count:", err)
		return err
	}

	if count > 0 {
		log.Println("Projects already exist in the database")
		return nil
	}

	//existingID, _ := primitive.ObjectIDFromHex("67386650a0d21b3a8f823722") // Known ObjectID as string

	projects := []interface{}{
		domain.Project{
			Id:             primitive.NewObjectID(),
			Name:           "Project Alpha",
			CompletionDate: time.Now().AddDate(0, 3, 0),
			MinMembers:     2,
			MaxMembers:     5,
			Manager:        domain.User{Id: "67386650a0d21b3a8f823722", Username: "alicej", Role: "Manager"},
		},
		domain.Project{
			Id:             primitive.NewObjectID(),
			Name:           "Project Beta",
			CompletionDate: time.Now().AddDate(0, 6, 0),
			MinMembers:     3,
			MaxMembers:     6,
			Manager:        domain.User{Id: "67386650a0d21b3a8f823722", Username: "alicej", Role: "Manager"},
		},
		domain.Project{
			Id:             primitive.NewObjectID(),
			Name:           "Project Gamma",
			CompletionDate: time.Now().AddDate(0, 1, 0),
			MinMembers:     1,
			MaxMembers:     4,
			Manager:        domain.User{Id: "67386650a0d21b3a8f823722", Username: "alicej", Role: "Manager"},
		},
		domain.Project{
			Id:             primitive.NewObjectID(),
			Name:           "Project Delta",
			CompletionDate: time.Now().AddDate(0, 9, 0),
			MinMembers:     4,
			MaxMembers:     10,
			Manager:        domain.User{Id: "67386650a0d21b3a8f823722", Username: "alicej", Role: "Manager"},
		},
	}

	// Insert initial projects
	_, err = collection.InsertMany(context.Background(), projects)
	if err != nil {
		log.Println("Error inserting initial projects:", err)
		return err
	}

	log.Println("Initial projects inserted successfully")
	return nil
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

	if err := insertInitialProjects(client); err != nil {
		log.Printf("Failed to insert initial projects: %v", err)
	}

	return &ProjectRepo{
		cli: client,
	}, nil
}

func (pr *ProjectRepo) Create(project *proto.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Provera CompletionDate pre poziva AsTime()
	if project.CompletionDate == nil {
		log.Println("CompletionDate is nil")
		return errors.New("completionDate is required")
	}

	completionDate := project.CompletionDate.AsTime()

	// Konvertuj proto.Project u domain.Project
	prj := &domain.Project{
		Name:           project.Name,
		CompletionDate: completionDate.UTC(), // Osigurajte da je u UTC
		MinMembers:     project.MinMembers,
		MaxMembers:     project.MaxMembers,
		Manager: domain.User{
			Id:       project.Manager.Id,
			Username: project.Manager.Username,
			Role:     project.Manager.Role,
		},
	}

	// Ubaci konvertovani domain.Project u MongoDB
	projectsCollection := pr.getCollection()

	result, err := projectsCollection.InsertOne(ctx, prj)
	if err != nil {
		log.Printf("Error inserting document: %v\n", err)
		return err
	}

	log.Printf("Document ID: %v\n", result.InsertedID)
	return nil
}

func (pr *ProjectRepo) GetAllProjects(id string) (domain.Projects, error) {
	// Initialize context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()
	var projects domain.Projects

	// Query only projects where the manager's ID matches the provided id
	filter := bson.M{"manager.username": id}
	cursor, err := projectsCollection.Find(ctx, filter)
	if err != nil {
		log.Println("Error finding projects:", err)
		return nil, err
	}

	// Decode the results into the projects slice
	if err = cursor.All(ctx, &projects); err != nil {
		log.Println("Error decoding projects:", err)
		return nil, err
	}

	return projects, nil
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

func (pr *ProjectRepo) GetById(id string) (*domain.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	// Convert id string to ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid ID format:", err)
		return nil, err
	}

	// Find project by _id
	filter := bson.M{"_id": objID}
	var project domain.Project
	err = projectsCollection.FindOne(ctx, filter).Decode(&project)
	if err != nil {
		log.Println("Error finding project by ID:", err)
		return nil, err
	}

	return &project, nil
}
