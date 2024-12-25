package repositories

import (
	"analytics-service/models"
	"context"
	"log"
	"os"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/trace"
)

type AnalyticsRepo struct {
	session *gocql.Session
	logger  *log.Logger
	Tracer  trace.Tracer
}

// NewAnalyticsRepo initializes a new AnalyticsRepo
func NewAnalyticsRepo(logger *log.Logger, tracer trace.Tracer) (*AnalyticsRepo, error) {
	db := os.Getenv("CASS_DB")
	// Connect to Cassandra cluster
	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	// Create 'analytics' keyspace
	err = session.Query(
		`CREATE KEYSPACE IF NOT EXISTS analytics
		WITH replication = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}`).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	// Connect to 'analytics' keyspace
	cluster.Keyspace = "analytics"
	cluster.Consistency = gocql.Quorum
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	return &AnalyticsRepo{
		session: session,
		logger:  logger,
		Tracer:  tracer,
	}, nil
}

// CloseSession closes the Cassandra session
func (ar *AnalyticsRepo) CloseSession() {
	ar.session.Close()
}

// CreateTables creates necessary tables in Cassandra
func (ar *AnalyticsRepo) CreateTables(ctx context.Context) {
	ctx, span := ar.Tracer.Start(ctx, "r.createTables")
	defer span.End()

	err := ar.session.Query(
		`CREATE TABLE IF NOT EXISTS analytics (
			project_id TEXT PRIMARY KEY,
			total_tasks INT,
			status_counts MAP<TEXT, INT>,
			task_status_durations MAP<TEXT, LIST<FROZEN<TaskStatusDuration>>>,
			member_tasks MAP<TEXT, LIST<TEXT>>,
			finished_early BOOLEAN
		)`).Exec()

	if err != nil {
		ar.logger.Println(err)
	}
}

// InsertAnalytics inserts an analytic entry into Cassandra
func (ar *AnalyticsRepo) InsertAnalytics(ctx context.Context, analytic *models.Analytic) error {
	ctx, span := ar.Tracer.Start(ctx, "r.insertAnalytics")
	defer span.End()

	err := ar.session.Query(
		`INSERT INTO analytics (project_id, total_tasks, status_counts, task_status_durations, member_tasks, finished_early) 
		VALUES (?, ?, ?, ?, ?, ?)`,

		analytic.ProjectID,
		analytic.TotalTasks,
		analytic.StatusCounts,
		convertTaskStatusDurationsToCassandra(analytic.TaskStatusDurations), // Convert to appropriate format
		convertMemberTasksToCassandra(analytic.MemberTasks),                 // Convert to appropriate format
		analytic.FinishedEarly,
	).Exec()

	if err != nil {
		ar.logger.Println(err)
		return err
	}
	return nil
}

// GetAnalyticsByProject retrieves an analytic entry by project ID
func (ar *AnalyticsRepo) GetAnalyticsByProject(ctx context.Context, projectID string) (*models.Analytic, error) {
	ctx, span := ar.Tracer.Start(ctx, "r.getAnalyticsByProject")
	defer span.End()

	var analytic models.Analytic
	var statusCounts map[string]int32
	var taskStatusDurations map[string][]models.TaskStatusDuration
	var memberTasks map[string][]string

	err := ar.session.Query(
		`SELECT project_id, total_tasks, status_counts, task_status_durations, member_tasks, finished_early 
		FROM analytics WHERE project_id = ?`,
		projectID,
	).Scan(
		&analytic.ProjectID,
		&analytic.TotalTasks,
		&statusCounts,
		&taskStatusDurations,
		&memberTasks,
		&analytic.FinishedEarly,
	)

	if err != nil {
		ar.logger.Println(err)
		return nil, err
	}

	// Correct the mapping for TaskStatusDurations and MemberTasks
	analytic.StatusCounts = statusCounts
	analytic.TaskStatusDurations = convertCassandraTaskDurationsToModel(taskStatusDurations)
	analytic.MemberTasks = convertCassandraMemberTasksToModel(memberTasks)

	return &analytic, nil
}

// Helper function to convert TaskStatusDurations to the appropriate Cassandra format
func convertTaskStatusDurationsToCassandra(taskStatusDurations map[string]models.TaskDurations) map[string][]models.TaskStatusDuration {
	result := make(map[string][]models.TaskStatusDuration)
	for taskID, taskDurations := range taskStatusDurations {
		result[taskID] = taskDurations.StatusDurations // Directly use the slice from TaskDurations
	}
	return result
}

// Helper function to convert MemberTasks to the appropriate Cassandra format
func convertMemberTasksToCassandra(memberTasks map[string]models.TaskAssignments) map[string][]string {
	result := make(map[string][]string)
	for memberID, assignments := range memberTasks {
		result[memberID] = assignments.Tasks
	}
	return result
}

// Helper function to convert Cassandra task status durations back to the model
func convertCassandraTaskDurationsToModel(cassandraDurations map[string][]models.TaskStatusDuration) map[string]models.TaskDurations {
	result := make(map[string]models.TaskDurations)
	for taskID, durations := range cassandraDurations {
		result[taskID] = models.TaskDurations{
			TaskID:          taskID,
			StatusDurations: durations,
		}
	}
	return result
}

// Helper function to convert Cassandra member tasks back to the model
func convertCassandraMemberTasksToModel(cassandraTasks map[string][]string) map[string]models.TaskAssignments {
	result := make(map[string]models.TaskAssignments)
	for memberID, tasks := range cassandraTasks {
		result[memberID] = models.TaskAssignments{
			Tasks: tasks,
		}
	}
	return result
}
