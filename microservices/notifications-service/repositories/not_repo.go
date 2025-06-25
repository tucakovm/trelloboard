package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net"
	"not_module/domain"
	"os"
	"strconv"
	"time"
)

type NotRepo struct {
	session *gocql.Session
	logger  *log.Logger
	Tracer  trace.Tracer
	cache   *redis.Client
}

func New(logger *log.Logger, tracer trace.Tracer) (*NotRepo, error) {
	cassandraHost := os.Getenv("CASSANDRA_HOST")
	cassandraPortStr := os.Getenv("CASSANDRA_PORT")
	port, err := strconv.Atoi(cassandraPortStr)
	if err != nil {
		logger.Println("Invalid CASSANDRA_PORT:", err)
		return nil, err
	}

	cluster := gocql.NewCluster(cassandraHost)
	cluster.Port = port
	cluster.Keyspace = "system"
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.Quorum

	systemSession, err := cluster.CreateSession()
	if err != nil {
		logger.Println("Error connecting to Cassandra system keyspace:", err)
		return nil, err
	}
	defer systemSession.Close()

	err = systemSession.Query(`
		CREATE KEYSPACE IF NOT EXISTS user
		WITH replication = {
			'class' : 'SimpleStrategy',
			'replication_factor' : 1
		}`).Exec()
	if err != nil {
		logger.Println("Error creating keyspace:", err)
		return nil, err
	}

	cluster.Keyspace = "user"
	userSession, err := cluster.CreateSession()
	if err != nil {
		logger.Println("Error connecting to 'user' keyspace:", err)
		return nil, err
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddr := net.JoinHostPort(redisHost, redisPort)

	clientRedis := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &NotRepo{
		session: userSession,
		logger:  logger,
		Tracer:  tracer,
		cache:   clientRedis,
	}, nil
}

// Disconnect from database
func (nr *NotRepo) CloseSession() {
	nr.session.Close()
}

func (nr *NotRepo) InitDB(ctx context.Context) {

	ctx, span := nr.Tracer.Start(ctx, "r.initNotDB")
	defer span.End()

	notifs := []*domain.Notification{
		{
			UserId:    "67386650a0d21b3a8f823722",
			CreatedAt: time.Now().AddDate(0, 0, -1),
			Message:   "not1.",
			Status:    "unread",
		},
		{
			UserId:    "67386650a0d21b3a8f823722",
			CreatedAt: time.Now().AddDate(0, 0, -2),
			Message:   "not2.",
			Status:    "unread",
		},
		{
			UserId:    "67386650a0d21b3a8f823722",
			CreatedAt: time.Now(),
			Message:   "not3.",
			Status:    "unread",
		},
	}

	for _, not := range notifs {
		nr.InsertNotByUser(ctx, not)
		log.Println("Inserted nots :")
		log.Println(not)
	}
}
func (nr *NotRepo) CreateTables(ctx context.Context) {
	ctx, span := nr.Tracer.Start(ctx, "r.createTables")
	defer span.End()
	err := nr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
        (user_id TEXT, 
        created_at TIMESTAMP, 
        not_id UUID, 
        message TEXT, 
        status TEXT,
        PRIMARY KEY ((user_id), not_id , created_at)) 
        WITH CLUSTERING ORDER BY (not_id ASC, created_at DESC)`,
			"not_by_user")).Exec()

	if err != nil {
		nr.logger.Println(err)
	}
	nr.InitDB(ctx)
}

func (nr *NotRepo) GetNotsByUser(ctx context.Context, id string) (domain.Notifications, error) {
	ctx, span := nr.Tracer.Start(ctx, "r.getAllNotsUser")
	defer span.End()
	scanner := nr.session.Query(`SELECT user_id, created_at, not_id, message, status FROM not_by_user WHERE user_id = ?`,
		id).Iter().Scanner()

	var nots domain.Notifications
	for scanner.Next() {
		var not domain.Notification
		err := scanner.Scan(&not.UserId, &not.CreatedAt, &not.NotificationId, &not.Message, &not.Status)
		if err != nil {
			nr.logger.Println(err)
			return nil, err
		}
		nots = append(nots, &not)
	}
	if err := scanner.Err(); err != nil {
		nr.logger.Println(err)
		return nil, err
	}

	log.Println("Repo nots !1!!!111: ")
	log.Println(nots)
	return nots, nil
}

func (nr *NotRepo) InsertNotByUser(ctx context.Context, not *domain.Notification) error {
	ctx, span := nr.Tracer.Start(ctx, "r.insertNotByUser")
	defer span.End()
	notId, _ := gocql.RandomUUID()
	err := nr.session.Query(
		`INSERT INTO not_by_user (user_id, created_at , not_id, message, status) 
		VALUES (?, ?, ?, ?, ?)`,
		not.UserId, not.CreatedAt, notId, not.Message, not.Status).Exec()
	if err != nil {
		nr.logger.Println(err)
		return err
	}
	log.Println("Inserted not :")
	log.Println(not)
	return nil
}

// REDIS -----------------------

func (nr *NotRepo) PostAll(id string, nots domain.Notifications, ctx context.Context) error {
	ctx, span := nr.Tracer.Start(ctx, "r.PostAllNotsCache")
	defer span.End()

	value, err := json.Marshal(nots)
	if err != nil {
		return err
	}

	err = nr.cache.Set(constructKeyProjects(id), value, 5*time.Second).Err()
	log.Println("Cache hit [PostAll]")
	return err
}

func (nr *NotRepo) GetAllNotsCache(username string, ctx context.Context) (domain.Notifications, error) {
	ctx, span := nr.Tracer.Start(ctx, "r.GetAllNotsCache")
	defer span.End()
	values, err := nr.cache.Get(constructKeyProjects(username)).Bytes()
	if err != nil {
		return domain.Notifications{}, err
	}

	products := &domain.Notifications{}
	err = json.Unmarshal(values, products)
	if err != nil {
		return domain.Notifications{}, err
	}

	//pr.logger.Println("Cache hit")
	log.Printf("Cache hit[GetAll]")
	return *products, nil
}
