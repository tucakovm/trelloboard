package repositories

import (
	"context"
	"fmt"
	"log"
	"os"
	"projects_module/domain"
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
			Members: []domain.User{
				{Id: "67386650a0d21b3a8f823720", Username: "bobsmith", Role: "User"},
			},
			Manager: domain.User{Id: "67386650a0d21b3a8f823722", Username: "alicej", Role: "Manager"},
		},
		domain.Project{
			Id:             primitive.NewObjectID(),
			Name:           "Project Beta",
			CompletionDate: time.Now().AddDate(0, 6, 0),
			MinMembers:     3,
			MaxMembers:     6,
			Members: []domain.User{
				{Id: "67386650a0d21b3a8f823720", Username: "bobsmith", Role: "User"},
			},
			Manager: domain.User{Id: "67386650a0d21b3a8f823722", Username: "alicej", Role: "Manager"},
		},
		domain.Project{
			Id:             primitive.NewObjectID(),
			Name:           "Project Gamma",
			CompletionDate: time.Now().AddDate(0, 1, 0),
			MinMembers:     1,
			MaxMembers:     4,
			Members:        []domain.User{},
			Manager:        domain.User{Id: "67386650a0d21b3a8f823722", Username: "alicej", Role: "Manager"},
		},
		domain.Project{
			Id:             primitive.NewObjectID(),
			Name:           "Project Delta",
			CompletionDate: time.Now().AddDate(0, 9, 0),
			MinMembers:     4,
			MaxMembers:     10,
			Members:        []domain.User{},
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

func (pr *ProjectRepo) Create(project *domain.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ubaci konvertovani domain.Project u MongoDB
	projectsCollection := pr.getCollection()

	result, err := projectsCollection.InsertOne(ctx, project)
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

func (pr *ProjectRepo) DoesManagerExistOnProject(id string) (bool, error) {
	// Initialize context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	// Query to find a project where the manager's ID matches the provided id
	filter := bson.M{"manager.username": id}
	count, err := projectsCollection.CountDocuments(ctx, filter)
	if err != nil {
		log.Println("Error counting projects:", err)
		return false, err
	}

	return count > 0, nil
}

func (pr *ProjectRepo) DoesUserExistOnProject(id string) (bool, error) {
	// Initialize context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()
	log.Println("Does user exist on project : " + id)
	// Query to find a project where the manager's ID matches the provided id
	filter := bson.M{"members.username": id}
	count, err := projectsCollection.CountDocuments(ctx, filter)
	if err != nil {
		log.Println("Error counting projects:", err)
		return false, err
	}

	return count > 0, nil
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

// AddMember adds a user to the project's member list
func (pr *ProjectRepo) AddMember(projectId string, user domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	// Convert projectId string to ObjectID
	objID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		log.Println("Invalid project ID format:", err)
		return err
	}

	// Update the project by appending the user to the Members array
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$push": bson.M{"members": user},
	}

	result, err := projectsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Error adding member to project:", err)
		return err
	}

	if result.ModifiedCount == 0 {
		log.Println("No project found with the given ID")
		return fmt.Errorf("no project found with the given ID")
	}

	log.Printf("User %s added to project %s", user.Username, projectId)
	return nil
}
func (pr *ProjectRepo) RemoveMember(projectId string, userId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()
	log.Printf("Usao u repo od remove membera")

	objID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		log.Println("Invalid project ID format:", err)
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$pull": bson.M{"members": bson.M{"_id": userId}},
	}

	result, err := projectsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Error removing member from project:", err)
		return err
	}

	if result.ModifiedCount == 0 {
		log.Println("No project found with the given ID or user not in the members list")
		return fmt.Errorf("no project found with the given ID or user not in the members list")
	}

	log.Printf("User %s removed from project %s", userId, projectId)
	return nil
}
