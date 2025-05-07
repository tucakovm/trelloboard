package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/go-redis/redis"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"projects_module/domain"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ProjectRepo struct {
	cli    *mongo.Client
	Tracer trace.Tracer
	cache  *redis.Client
	esdb   *ESDBEventStream
}

type ESDBEventStream struct {
	client *esdb.Client
	group  string
	sub    *esdb.PersistentSubscription
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
				{Id: "6749ac3cfc079b8c923bb9d5", Username: "bobsmith", Role: "User"},
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
				{Id: "6749ac3cfc079b8c923bb9d5", Username: "bobsmith", Role: "User"},
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

func New(ctx context.Context, clientESDB *esdb.Client, group string, logger *log.Logger, tracer trace.Tracer) (*ProjectRepo, error) {

	dburi := os.Getenv("MONGO_DB_URI")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	clientRedis := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	// Optionally, check if the connection is valid by pinging the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	opts := esdb.PersistentAllSubscriptionOptions{
		From: esdb.Start{},
	}
	err = clientESDB.CreatePersistentSubscriptionAll(context.Background(), group, opts)
	if err != nil {
		// persistent subscription group already exists
		log.Println(err)
	}
	eventStream := &ESDBEventStream{
		client: clientESDB,
		group:  group,
	}
	err = eventStream.subscribe()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := insertInitialProjects(client); err != nil {
		log.Printf("Failed to insert initial projects: %v", err)
	}

	return &ProjectRepo{
		cli:    client,
		Tracer: tracer,
		cache:  clientRedis,
		esdb:   eventStream,
	}, nil
}

func (pr *ProjectRepo) Create(project *domain.Project, ctx context.Context) error {
	ctx, span := pr.Tracer.Start(ctx, "r.createProject")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ubaci konvertovani domain.Project u MongoDB
	projectsCollection := pr.getCollection()

	result, err := projectsCollection.InsertOne(ctx, project)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error inserting document: %v\n", err)
		return err
	}

	log.Printf("Document ID: %v\n", result.InsertedID)
	return nil
}

func (pr *ProjectRepo) GetAllProjects(id string, ctx context.Context) (domain.Projects, error) {
	if pr.Tracer == nil {
		log.Println("Service is nil")
		return nil, status.Error(codes.Internal, "service is not initialized")
	}

	ctx, span := pr.Tracer.Start(ctx, "r.getUserByEmail")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()
	var projects domain.Projects

	filter := bson.M{
		"$or": []bson.M{
			{"manager.username": id},
			{"members.username": id},
		},
	}

	cursor, err := projectsCollection.Find(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	if err = cursor.All(ctx, &projects); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	return projects, nil
}

func (pr *ProjectRepo) DoesManagerExistOnProject(id string, ctx context.Context) (bool, error) {
	// Initialize context (after 5 seconds timeout, abort operation)

	ctx, span := pr.Tracer.Start(ctx, "r.doesManagerExistOnProject")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	// Query to find a project where the manager's ID matches the provided id
	filter := bson.M{"manager.username": id}
	count, err := projectsCollection.CountDocuments(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error counting projects:", err)
		return false, err
	}

	return count > 0, nil
}

func (pr *ProjectRepo) DoesUserExistOnProject(id string, ctx context.Context) (bool, error) {
	// Initialize context (after 5 seconds timeout, abort operation)

	ctx, span := pr.Tracer.Start(ctx, "r.DoesUserExistOnProject")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()
	log.Println("Does user exist on project : " + id)
	// Query to find a project where the manager's ID matches the provided id
	filter := bson.M{"members.username": id}
	count, err := projectsCollection.CountDocuments(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error counting projects:", err)
		return false, err
	}

	return count > 0, nil
}

func (pr *ProjectRepo) DoesMemberExistOnProject(projectId string, userId string, ctx context.Context) (bool, error) {

	ctx, span := pr.Tracer.Start(ctx, "r.doesMemberExistOnProject")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	objID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Invalid project ID: %v\n", err)
		return false, err
	}

	filter := bson.M{
		"_id":              objID,
		"members.username": userId,
	}

	var project bson.M
	err = projectsCollection.FindOne(ctx, filter).Decode(&project)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		log.Printf("Error querying project: %v\n", err)
		return false, err
	}

	return true, nil
}

func (pr *ProjectRepo) Delete(id string, ctx context.Context) (*domain.Project, error) {
	ctx, span := pr.Tracer.Start(ctx, "r.deleteProject")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println(err)
		return nil, err
	}

	var project domain.Project
	err = projectsCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("Project not found")
			return nil, nil
		}
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println(err)
		return nil, err
	}

	result, err := projectsCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println(err)
		return nil, err
	}

	log.Printf("Documents deleted: %v\n", result.DeletedCount)
	return &project, nil
}

func (pr *ProjectRepo) GetById(id string, ctx context.Context) (*domain.Project, error) {
	ctx, span := pr.Tracer.Start(ctx, "r.getProjectById")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	// Convert id string to ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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
func (pr *ProjectRepo) AddMember(projectId string, user domain.User, ctx context.Context) error {
	ctx, span := pr.Tracer.Start(ctx, "r.addMemberToProject")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	// Convert projectId string to ObjectID
	objID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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
		span.SetStatus(otelCodes.Error, err.Error())
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

func (pr *ProjectRepo) RemoveMember(projectId string, userId string, ctx context.Context) error {

	ctx, span := pr.Tracer.Start(ctx, "r.removeMemberFromProject")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()
	log.Printf("Usao u repo od remove membera")

	objID, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Invalid project ID format:", err)
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$pull": bson.M{"members": bson.M{"_id": userId}},
	}

	result, err := projectsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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

func (pr *ProjectRepo) MarkAsDeleting(id string, ctx context.Context) error {
	ctx, span := pr.Tracer.Start(ctx, "r.markProjectAsDeleting")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Invalid ID format:", err)
		return err
	}

	var project struct {
		Name string `bson:"name"`
	}
	err = projectsCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&project)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error finding project:", err)
		return err
	}

	const suffix = " (In deletion process)"
	if !strings.HasSuffix(project.Name, suffix) {
		project.Name += suffix
	}

	update := bson.M{
		"$set": bson.M{"name": project.Name},
	}

	_, err = projectsCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error updating project name:", err)
		return err
	}

	return nil
}

func (pr *ProjectRepo) UnmarkAsDeleting(id string, ctx context.Context) error {
	ctx, span := pr.Tracer.Start(ctx, "r.unmarkProjectAsDeleting")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := pr.getCollection()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Invalid ID format:", err)
		return err
	}

	var project struct {
		Name string `bson:"name"`
	}
	err = projectsCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&project)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error finding project:", err)
		return err
	}

	const suffix = " (In deletion process)"
	if strings.HasSuffix(project.Name, suffix) {
		project.Name = strings.TrimSuffix(project.Name, suffix)
	}

	update := bson.M{
		"$set": bson.M{"name": project.Name},
	}

	_, err = projectsCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error updating project name:", err)
		return err
	}

	return nil
}

// REDIS -------------------------------------------------------------------->

func (pc *ProjectRepo) Post(project *domain.Project, ctx context.Context) error {
	ctx, span := pc.Tracer.Start(ctx, "r.PostProjectCache")
	defer span.End()

	key := project.Manager.Username
	value, err := json.Marshal(project)
	if err != nil {
		return err
	}
	tllKey := constructKeyProjects(key)
	ttl, err := pc.cache.TTL(tllKey).Result()
	if err != nil {
		return err
	}

	err = pc.cache.Set(constructKeyProjects(key), value, ttl).Err()

	return nil
}

func (pc *ProjectRepo) PostOne(project *domain.Project, ctx context.Context) error {
	ctx, span := pc.Tracer.Start(ctx, "r.PostProjectCache")
	defer span.End()

	prjID := project.Id
	value, err := json.Marshal(project)
	if err != nil {
		return err
	}

	err = pc.cache.Set(constructKeyOneProject(prjID.Hex()), value, 30*time.Second).Err()
	log.Println("Cache hit [Post]")

	return err
}

func (pc *ProjectRepo) PostAll(managerId string, products domain.Projects, ctx context.Context) error {
	ctx, span := pc.Tracer.Start(ctx, "r.PostAllProjectsCache")
	defer span.End()

	value, err := json.Marshal(products)
	if err != nil {
		return err
	}

	err = pc.cache.Set(constructKeyProjects(managerId), value, 30*time.Second).Err()
	log.Println("Cache hit [PostAll]")
	return err
}

func (pc *ProjectRepo) Get(projectId string, ctx context.Context) (*domain.Project, error) {
	ctx, span := pc.Tracer.Start(ctx, "r.GetByIdCache")
	defer span.End()

	value, err := pc.cache.Get(constructKeyOneProject(projectId)).Bytes()
	if err != nil {
		return nil, err
	}

	product := &domain.Project{}
	err = json.Unmarshal(value, product)
	if err != nil {
		return nil, err
	}

	//pc.logger.Println("Cache hit")
	//log.Printf("Cache hit[Get]")
	return product, nil
}

func (pr *ProjectRepo) GetAllProjectsCache(managerId string, ctx context.Context) (domain.Projects, error) {
	ctx, span := pr.Tracer.Start(ctx, "r.GetAllProjectsCache")
	defer span.End()
	values, err := pr.cache.Get(constructKeyProjects(managerId)).Bytes()
	if err != nil {
		return domain.Projects{}, err
	}

	products := &domain.Projects{}
	err = json.Unmarshal(values, products)
	if err != nil {
		return domain.Projects{}, err
	}

	//pr.logger.Println("Cache hit")
	//log.Printf("Cache hit[GetAll]")
	return *products, nil
}

func (pc *ProjectRepo) DeleteByKey(key string, username string, ctx context.Context) error {
	ctx, span := pc.Tracer.Start(ctx, "r.DeleteByKey")
	defer span.End()

	err := pc.cache.Del(key).Err()
	if err != nil {
		log.Printf("Failed to delete cache for key %s: %v", key, err)
	}

	values, err := pc.cache.Get(constructKeyProjects(username)).Bytes()
	if err != nil {
		return err
	}

	projects := &domain.Projects{}
	err = json.Unmarshal(values, projects)
	if err != nil {
		return err
	}

	filteredProjects := domain.Projects{}

	for _, prj := range *projects {
		if prj.Id.Hex() != key {
			filteredProjects = append(filteredProjects, prj)
		}
	}

	err = pc.cache.Del(constructKeyProjects(username)).Err()
	if err != nil {
		log.Printf("Failed to delete cache projects for key %s: %v", key, err)
	}

	value, err := json.Marshal(filteredProjects)
	if err != nil {
		return err
	}

	err = pc.cache.Set(username, value, 30*time.Second).Err()
	if err != nil {
		log.Printf("Failed to set cache projects for key %s: %v", key, err)
		return err
	}

	return nil
}

//func (pr *ProjectRepo) Exists(id string, managerId string) bool {
//	cnt, err := pr.cache.Exists(constructKeyOneProject(id, managerId)).Result()
//	if cnt == 1 {
//		return true
//	}
//	if err != nil {
//		return false
//	}
//	return false
//}

// ESDB --------------------------------------------------->

func (s *ESDBEventStream) subscribe() error {
	opts := esdb.ConnectToPersistentSubscriptionOptions{}
	sub, err := s.client.ConnectToPersistentSubscriptionToAll(context.Background(), s.group, opts)
	if err != nil {
		return err
	}
	s.sub = sub
	return nil
}

func (r *ProjectRepo) AppendEvent(ctx context.Context, streamID string, data []byte, eventType string) error {
	ctx, span := r.Tracer.Start(ctx, "r.esdb-AppendEvent")
	defer span.End()

	event := esdb.EventData{
		ContentType: esdb.JsonContentType,
		EventType:   eventType,
		Data:        data,
	}

	opts := esdb.AppendToStreamOptions{
		ExpectedRevision: esdb.Any{},
	}

	stream := "project-" + streamID

	_, err := r.esdb.client.AppendToStream(ctx, stream, opts, event)
	if err != nil {
		return fmt.Errorf("failed to append event to stream %s: %w", stream, err)
	}

	return nil
}
