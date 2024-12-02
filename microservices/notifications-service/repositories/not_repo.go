package repositories

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/trace"
	"log"
	"not_module/domain"
	"os"
	"time"
)

type NotRepo struct {
	session *gocql.Session
	logger  *log.Logger
	Tracer  trace.Tracer
}

func New(logger *log.Logger, tracer trace.Tracer) (*NotRepo, error) {
	db := os.Getenv("CASS_DB")
	// Connect to default keyspace
	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	// Create 'student' keyspace
	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "user", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	// Connect to student keyspace
	cluster.Keyspace = "user"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	repo := &NotRepo{
		session: session,
		logger:  logger,
		Tracer:  tracer,
	}

	return repo, nil
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
