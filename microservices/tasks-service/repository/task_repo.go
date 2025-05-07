package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EventStore/EventStore-Client-Go/esdb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"tasks-service/domain"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskRepo struct {
	Cli    *mongo.Client
	Tracer trace.Tracer
	cache  *redis.Client
	esdb   *ESDBEventStream
}

type ESDBEventStream struct {
	client *esdb.Client
	group  string
	sub    *esdb.PersistentSubscription
}

func NewTaskRepo(ctx context.Context, clientESDB *esdb.Client, group string, logger *log.Logger, tracer trace.Tracer) (*TaskRepo, error) {
	// MongoDB URI
	dburi := os.Getenv("MONGO_DB_URI")
	if dburi == "" {
		return nil, fmt.Errorf("MONGO_DB_URI is not set")
	}

	// Redis connection
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		logger.Printf("Failed to connect to MongoDB: %v", err)
		return nil, err
	}

	// Check MongoDB connection
	if err := client.Ping(ctx, nil); err != nil {
		logger.Printf("MongoDB ping failed: %v", err)
		return nil, err
	}
	logger.Println("Connected to MongoDB successfully")

	// Setup Redis client
	clientRedis := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	// Initialize EventStoreDB stream
	eventStream := &ESDBEventStream{
		client: clientESDB,
		group:  group,
	}

	// Optional: subscribe to a persistent subscription, if needed
	if err := eventStream.subscribe(); err != nil {
		logger.Printf("Failed to subscribe to ESDB: %v", err)
		return nil, err
	}

	// Insert initial tasks into MongoDB (optional)
	if err := insertInitialTasks(client); err != nil {
		logger.Printf("Failed to insert initial tasks: %v", err)
	}

	// Return initialized TaskRepo
	return &TaskRepo{
		Cli:    client,
		Tracer: tracer,
		cache:  clientRedis,
		esdb:   eventStream,
	}, nil
}

func (tr *TaskRepo) Disconnect(ctx context.Context) error {
	err := tr.Cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func insertInitialTasks(client *mongo.Client) error {
	collection := client.Database("mongoDemo").Collection("tasks")
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println("Error checking task count:", err)
		return err
	}

	if count > 0 {
		log.Println("Tasks already exist in the database")
		return nil
	}

	// Define initial tasks to insert
	tasks := []interface{}{
		domain.Task{
			Name:        "Task 1",
			Description: "This is the first task.",
			Status:      domain.Status(0),
			ProjectID:   "jnasdndslksad",
		},
		domain.Task{
			Name:        "Task 2",
			Description: "This is the second task.",
			Status:      domain.Status(0),
			ProjectID:   "lksaddsmamkls",
		},
	}

	// Insert initial tasks
	_, err = collection.InsertMany(context.Background(), tasks)
	if err != nil {
		log.Println("Error inserting initial tasks:", err)
		return err
	}

	log.Println("Initial tasks inserted successfully")
	return nil
}

func (tr *TaskRepo) getCollection() *mongo.Collection {
	if tr.Cli == nil {
		log.Println("Mongo client is nil!")
		return nil
	}

	if err := tr.Cli.Ping(context.Background(), nil); err != nil {
		log.Println("Error pinging MongoDB, connection lost:", err)
		return nil
	}

	return tr.Cli.Database("mongoDemo").Collection("tasks")
}

func (tr *TaskRepo) Create(task domain.Task, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.createTask")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println(task)

	collection := tr.getCollection()
	if collection == nil {
		log.Println("Failed to retrieve collection")
		return fmt.Errorf("collection is nil")
	}

	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error inserting task:", err, task)
		return err
	}

	log.Println("Task created successfully:", task)
	return nil
}

func (tr *TaskRepo) GetAll() ([]domain.Task, error) {

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

	var tasks []domain.Task
	for cursor.Next(ctx) {
		var task domain.Task
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

func (tr *TaskRepo) Delete(id string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.deleteTask")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	taskCollection := tr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := taskCollection.DeleteOne(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println(err)
		return err
	}
	log.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}

func (tr *TaskRepo) DeleteAllByProjectID(projectID string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.deleteAllByProjectId")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	filter := bson.M{"project_id": projectID}
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error deleting tasks by ProjectID:", err)
		return err
	}

	log.Printf("Tasks with ProjectID %s deleted successfully", projectID)
	return nil
}

func (tr *TaskRepo) GetAllByProjectID(projectID string, ctx context.Context) (domain.Tasks, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.getAllByProjectId")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasksCollection := tr.getCollection()
	var tasks domain.Tasks

	// Query only tasks where the project_id matches the ObjectId
	filter := bson.M{"project_id": projectID}
	cursor, err := tasksCollection.Find(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error finding tasks:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor and decode each document into the tasks slice
	for cursor.Next(ctx) {
		var task *domain.Task
		if err := cursor.Decode(&task); err != nil {
			span.SetStatus(otelCodes.Error, err.Error())
			log.Println("Error decoding task:", err)
			continue
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Cursor error:", err)
		return nil, err
	}

	log.Printf("Fetched %d tasks with ProjectID %s", len(tasks), projectID)
	return tasks, nil
}

func (tr *TaskRepo) GetById(id string, ctx context.Context) (*domain.Task, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.getById")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := tr.getCollection()

	// Convert id string to ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Invalid ID format:", err)
		return nil, err
	}

	// Find project by _id
	filter := bson.M{"_id": objID}
	var t domain.Task
	err = projectsCollection.FindOne(ctx, filter).Decode(&t)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error finding task by ID:", err)
		return nil, err
	}

	return &t, nil
}

func (tr *TaskRepo) HasIncompleteTasksByProject(id string, ctx context.Context) (bool, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.hasIncompleteTasksByProject")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := tr.getCollection()

	// Convert id string to ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Invalid ID format:", err)
		return false, err
	}

	filter := bson.M{
		"project_id": objID,
		"status": bson.M{
			"$ne": "Done",
		},
	}

	// Check if there is at least one matching document
	count, err := projectsCollection.CountDocuments(ctx, filter)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error checking for incomplete tasks:", err)
		return false, err
	}
	log.Println(count)
	return count > 0, nil
}

func (tr *TaskRepo) AddMember(taskId string, user domain.User, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.addMember")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := tr.getCollection()

	objID, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Invalid project ID format:", err)
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$push": bson.M{"members": user},
	}

	result, err := projectsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error adding member to task:", err)
		return err
	}

	if result.ModifiedCount == 0 {
		log.Println("No task found with the given ID")
		return fmt.Errorf("no task found with the given ID")
	}

	return nil
}
func (tr *TaskRepo) RemoveMember(projectId string, userId string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.removeMember")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectsCollection := tr.getCollection()

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
		log.Println("Error removing member from task:", err)
		return err
	}

	if result.ModifiedCount == 0 {
		log.Println("No task found with the given ID or user not in the members list")
		return fmt.Errorf("no task found with the given ID or user not in the members list")
	}

	return nil
}
func (tr *TaskRepo) Update(task domain.Task, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.UpdateTask")
	defer span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	objID := task.Id

	update := bson.M{
		"$set": bson.M{
			"name":        task.Name,
			"description": task.Description,
			"members":     task.Members,
			"status":      task.Status,
		},
	}

	filter := bson.M{"_id": objID}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error updating task:", err)
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no task found with the given ID")
	}

	log.Printf("Updated task with ID: %s", task.Id)
	return nil
}

func (tr *TaskRepo) MarkTasksAsDeleting(projectID string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.markTasksAsDeleting")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	// Aggregation pipeline update
	update := mongo.Pipeline{
		{{"$set", bson.D{
			{"name", bson.D{
				{"$concat", bson.A{"$name", " (In deletion process)"}},
			}},
		}}},
	}

	_, err := collection.UpdateMany(ctx, bson.M{"project_id": projectID}, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error updating task names:", err)
		return err
	}

	log.Printf("Tasks for ProjectID %s marked as 'In deletion process'", projectID)
	return nil
}

func (tr *TaskRepo) UnmarkTasksAsDeleting(projectID string, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.unmarkTasksAsDeleting")
	defer span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := tr.getCollection()
	if collection == nil {
		return fmt.Errorf("failed to retrieve collection")
	}

	// Aggregation pipeline update
	update := mongo.Pipeline{
		{{"$set", bson.D{
			{"name", bson.D{
				{"$replaceOne", bson.D{
					{"input", "$name"},
					{"find", " (In deletion process)"},
					{"replacement", ""},
				}},
			}},
		}}},
	}

	_, err := collection.UpdateMany(ctx, bson.M{"project_id": projectID}, update)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error unmarking task names:", err)
		return err
	}

	log.Printf("Tasks for ProjectID %s unmarked from 'In deletion process'", projectID)
	return nil
}

//REDIS---------------

func (tr *TaskRepo) Post(task *domain.Task, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.PostProjectCache")
	defer span.End()

	key := task.ProjectID
	value, err := json.Marshal(task)
	if err != nil {
		return err
	}
	tllKey := constructKeyProjects(key)
	ttl, err := tr.cache.TTL(tllKey).Result()
	if err != nil {
		return err
	}

	err = tr.cache.Set(constructKeyProjects(key), value, ttl).Err()
	log.Println("cache set return:", err)

	return nil
}

func (tr *TaskRepo) PostOne(task *domain.Task, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.PostTaskCache")
	defer span.End()

	taskId := task.Id
	value, err := json.Marshal(task)
	if err != nil {
		return err
	}

	err = tr.cache.Set(constructKeyOneProject(taskId.Hex()), value, 5*time.Second).Err()
	log.Println("Cache hit [Post]")

	return err
}

func (tr *TaskRepo) PostAll(projectId string, task domain.Tasks, ctx context.Context) error {
	ctx, span := tr.Tracer.Start(ctx, "r.PostAllProjectsCache")
	defer span.End()

	value, err := json.Marshal(task)
	if err != nil {
		return err
	}

	err = tr.cache.Set(constructKeyProjects(projectId), value, 5*time.Second).Err()
	log.Println("Cache hit [PostAll]")
	return err
}

func (tr *TaskRepo) GetCache(projectId string, ctx context.Context) (*domain.Task, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.GetByIdCache")
	defer span.End()

	value, err := tr.cache.Get(constructKeyOneProject(projectId)).Bytes()
	if err != nil {
		return nil, err
	}

	product := &domain.Task{}
	err = json.Unmarshal(value, product)
	if err != nil {
		return nil, err
	}

	//pc.logger.Println("Cache hit")
	log.Printf("Cache hit[Get]")
	return product, nil
}

func (tr *TaskRepo) GetAllProjectsCache(managerId string, ctx context.Context) (domain.Tasks, error) {
	ctx, span := tr.Tracer.Start(ctx, "r.GetAllTasksCache")
	defer span.End()
	values, err := tr.cache.Get(constructKeyProjects(managerId)).Bytes()
	if err != nil {
		return domain.Tasks{}, err
	}

	products := &domain.Tasks{}
	err = json.Unmarshal(values, products)
	if err != nil {
		return domain.Tasks{}, err
	}

	//pr.logger.Println("Cache hit")
	log.Printf("Cache hit[GetAll]")
	return *products, nil
}

func (pc *TaskRepo) DeleteByKey(key string, username string, ctx context.Context) error {
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

	tasks := &domain.Tasks{}
	err = json.Unmarshal(values, tasks)
	if err != nil {
		return err
	}

	filteredTasks := domain.Tasks{}

	for _, t := range *tasks {
		if t.Id.Hex() != key {
			filteredTasks = append(filteredTasks, t)
		}
	}

	err = pc.cache.Del(constructKeyProjects(username)).Err()
	if err != nil {
		log.Printf("Failed to delete cache projects for key %s: %v", key, err)
	}

	value, err := json.Marshal(filteredTasks)
	if err != nil {
		return err
	}

	err = pc.cache.Set(username, value, 5*time.Second).Err()
	if err != nil {
		log.Printf("Failed to set cache projects for key %s: %v", key, err)
		return err
	}

	return nil
}

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

func (r *TaskRepo) AppendEvent(ctx context.Context, streamID string, data []byte, eventType string) error {
	log.Println("Append event repo")

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
