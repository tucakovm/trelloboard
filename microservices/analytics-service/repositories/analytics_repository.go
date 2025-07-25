package repositories

import (
	"analytics-service/models"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/trace"
)

type AnalyticsRepo struct {
	session *gocql.Session
	logger  *log.Logger
	Tracer  trace.Tracer
}

func (ar *AnalyticsRepo) InsertTestAnalytics(ctx context.Context, projectID string) error {
	testAnalytic := &models.Analytic{
		ProjectID:  projectID,
		TotalTasks: 2,
		StatusCounts: map[string]int32{
			"Pending": 2,
			"Working": 0,
			"Done":    0,
		},
		TaskStatusDurations: map[string]models.TaskDurations{
			"64bc8a5f1a2b3c4d5e6f7890": {
				TaskID: "64bc8a5f1a2b3c4d5e6f7890",
				StatusDurations: []models.TaskStatusDuration{
					{Status: "Pending", Duration: 24.0},
				},
			},
			"64bc8a6f1a2b3c4d5e6f7891": {
				TaskID: "64bc8a6f1a2b3c4d5e6f7891",
				StatusDurations: []models.TaskStatusDuration{
					{Status: "Pending", Duration: 12.0},
				},
			},
		},
		MemberTasks: map[string]models.TaskAssignments{
			"bobsmith": {Tasks: []string{"64bc8a5f1a2b3c4d5e6f7890"}},
		},
		FinishedEarly: false,
	}

	err := ar.InsertAnalytics(ctx, testAnalytic)
	if err != nil {
		ar.logger.Printf("Failed to insert test analytics: %v", err)
		return err
	}

	ar.logger.Println("Successfully inserted test analytics for project", projectID)
	return nil
}

func NewAnalyticsRepo(logger *log.Logger, tracer trace.Tracer) (*AnalyticsRepo, error) {
	dbHost := os.Getenv("CASS_DB")
	if dbHost == "" {
		logger.Println("CASS_DB environment variable is not set")
		return nil, gocql.ErrNoConnections
	}

	cluster := gocql.NewCluster(dbHost)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println("Failed to connect to Cassandra system keyspace:", err)
		return nil, err
	}

	err = session.Query(`
		CREATE KEYSPACE IF NOT EXISTS analytics 
		WITH replication = {'class':'SimpleStrategy', 'replication_factor':1}
	`).Exec()
	if err != nil {
		logger.Printf("Failed to create analytics keyspace: %v", err)
		session.Close()
		return nil, err
	}
	session.Close()

	cluster.Keyspace = "analytics"
	cluster.Consistency = gocql.Quorum
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println("Failed to connect to analytics keyspace:", err)
		return nil, err
	}

	return &AnalyticsRepo{
		session: session,
		logger:  logger,
		Tracer:  tracer,
	}, nil
}

func (ar *AnalyticsRepo) CloseSession() {
	ar.session.Close()
}

func (ar *AnalyticsRepo) CreateTables(ctx context.Context) {
	ctx, span := ar.Tracer.Start(ctx, "r.createTables")
	defer span.End()

	ar.logger.Println("Starting table and UDT creation")

	err := ar.session.Query(`
	CREATE TYPE IF NOT EXISTS TaskStatusDuration (
		status TEXT,
		duration DOUBLE,
		timestamp BIGINT
	)
`).Exec()
	if err != nil {
		ar.logger.Printf("Failed to create UDT TaskStatusDuration: %v", err)
		return
	}

	err = ar.session.Query(`
		CREATE TABLE IF NOT EXISTS analytics (
			project_id TEXT PRIMARY KEY,
			total_tasks INT,
			status_counts MAP<TEXT, INT>,
			task_status_durations MAP<TEXT, FROZEN<LIST<FROZEN<TaskStatusDuration>>>>,
			member_tasks MAP<TEXT, FROZEN<LIST<TEXT>>>,
			finished_early BOOLEAN
		)
	`).Exec()
	if err != nil {
		ar.logger.Printf("Failed to create analytics table: %v", err)
		return
	}

	ar.InsertTestAnalytics(ctx, "67386650a0d21b3a8f823723")

	ar.logger.Println("Successfully created/verified analytics table and UDT")
}

func (ar *AnalyticsRepo) InsertAnalytics(ctx context.Context, analytic *models.Analytic) error {
	ctx, span := ar.Tracer.Start(ctx, "r.insertAnalytics")
	defer span.End()

	err := ar.session.Query(`
		INSERT INTO analytics (project_id, total_tasks, status_counts, task_status_durations, member_tasks, finished_early)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		analytic.ProjectID,
		analytic.TotalTasks,
		analytic.StatusCounts,
		convertTaskStatusDurationsToCassandra(analytic.TaskStatusDurations),
		convertMemberTasksToCassandra(analytic.MemberTasks),
		analytic.FinishedEarly,
	).Exec()

	if err != nil {
		ar.logger.Printf("InsertAnalytics error: %v", err)
		return err
	}
	return nil
}

func (ar *AnalyticsRepo) GetAnalyticsByProject(ctx context.Context, projectID string) (*models.Analytic, error) {
	ctx, span := ar.Tracer.Start(ctx, "r.getAnalyticsByProject")
	defer span.End()

	m := map[string]interface{}{}

	iter := ar.session.Query(`
        SELECT project_id, total_tasks, status_counts, task_status_durations, member_tasks, finished_early
        FROM analytics WHERE project_id = ?`, projectID).Iter()

	if !iter.MapScan(m) {
		ar.logger.Println("No analytics found for project", projectID)
		return nil, nil
	}

	rawTaskStatusDurations := m["task_status_durations"].(map[string][]map[string]interface{})
	memberTasks := m["member_tasks"].(map[string][]string)

	statusCountsRaw := m["status_counts"].(map[string]int)
	statusCounts := make(map[string]int32)
	for k, v := range statusCountsRaw {
		statusCounts[k] = int32(v)
	}

	analytic := &models.Analytic{
		ProjectID:           m["project_id"].(string),
		TotalTasks:          int32(m["total_tasks"].(int)),
		StatusCounts:        statusCounts,
		TaskStatusDurations: enrichDurationsWithLiveTime(convertRawCassandraDurations(rawTaskStatusDurations)),
		MemberTasks:         convertCassandraMemberTasksToModel(memberTasks),
		FinishedEarly:       m["finished_early"].(bool),
	}

	return analytic, nil
}

func (ar *AnalyticsRepo) UpdateAnalytics(ctx context.Context, projectID string, analytic *models.Analytic) error {
	ctx, span := ar.Tracer.Start(ctx, "r.updateAnalytics")
	defer span.End()

	err := ar.session.Query(`
		UPDATE analytics SET
			total_tasks = ?,
			status_counts = ?,
			task_status_durations = ?,
			member_tasks = ?,
			finished_early = ?
		WHERE project_id = ?
	`,
		analytic.TotalTasks,
		analytic.StatusCounts,
		convertTaskStatusDurationsToCassandra(analytic.TaskStatusDurations),
		convertMemberTasksToCassandra(analytic.MemberTasks),
		analytic.FinishedEarly,
		projectID,
	).Exec()

	if err != nil {
		ar.logger.Printf("UpdateAnalytics error: %v", err)
		return err
	}
	return nil
}

func convertTaskStatusDurationsToCassandra(taskStatusDurations map[string]models.TaskDurations) map[string][]map[string]interface{} {
	result := make(map[string][]map[string]interface{})
	for taskID, taskDurations := range taskStatusDurations {
		var cassandraList []map[string]interface{}
		for _, sd := range taskDurations.StatusDurations {
			cassandraList = append(cassandraList, map[string]interface{}{
				"status":    sd.Status,
				"duration":  sd.Duration,
				"timestamp": sd.Timestamp,
			})
		}
		result[taskID] = cassandraList
	}
	return result
}

func convertMemberTasksToCassandra(memberTasks map[string]models.TaskAssignments) map[string][]string {
	result := make(map[string][]string)
	for memberID, assignments := range memberTasks {
		result[memberID] = assignments.Tasks
	}
	return result
}

func convertCassandraMemberTasksToModel(cassandraTasks map[string][]string) map[string]models.TaskAssignments {
	result := make(map[string]models.TaskAssignments)
	for memberID, tasks := range cassandraTasks {
		result[memberID] = models.TaskAssignments{Tasks: tasks}
	}
	return result
}

func convertRawCassandraDurations(raw map[string][]map[string]interface{}) map[string]models.TaskDurations {
	result := make(map[string]models.TaskDurations)
	for taskID, rawDurations := range raw {
		var durations []models.TaskStatusDuration
		for _, entry := range rawDurations {
			status, _ := entry["status"].(string)
			duration, _ := entry["duration"].(float64)
			timestampRaw, _ := entry["timestamp"]

			var timestamp int64
			switch ts := timestampRaw.(type) {
			case int:
				timestamp = int64(ts)
			case int64:
				timestamp = ts
			case float64:
				timestamp = int64(ts)
			}

			durations = append(durations, models.TaskStatusDuration{
				Status:    status,
				Duration:  duration,
				Timestamp: timestamp,
			})
		}
		result[taskID] = models.TaskDurations{
			TaskID:          taskID,
			StatusDurations: durations,
		}
	}
	return result
}

func (ar *AnalyticsRepo) AddNewTaskToAnalytics(ctx context.Context, projectID, taskID string) error {
	ctx, span := ar.Tracer.Start(ctx, "r.addNewTaskToAnalytics")
	defer span.End()

	analytic, err := ar.GetAnalyticsByProject(ctx, projectID)
	if err != nil {
		ar.logger.Printf("Failed to fetch analytics for project %s: %v", projectID, err)
		return err
	}
	if analytic == nil {
		ar.logger.Printf("No analytics record found for project %s", projectID)
		return fmt.Errorf("analytics not found for project %s", projectID)
	}

	analytic.TaskStatusDurations[taskID] = models.TaskDurations{
		TaskID: taskID,
		StatusDurations: []models.TaskStatusDuration{
			{
				Status:    "Pending",
				Duration:  0.0,
				Timestamp: time.Now().Unix(),
			},
		},
	}

	analytic.TotalTasks++
	analytic.StatusCounts["Pending"]++

	// Ažuriraj analytics u bazi
	if err := ar.UpdateAnalytics(ctx, projectID, analytic); err != nil {
		ar.logger.Printf("Failed to update analytics with new task for project %s: %v", projectID, err)
		return err
	}

	ar.logger.Printf("Successfully added new task %s to analytics for project %s", taskID, projectID)
	return nil
}

func enrichDurationsWithLiveTime(durations map[string]models.TaskDurations) map[string]models.TaskDurations {
	now := time.Now().Unix()

	for taskID, taskDur := range durations {
		if len(taskDur.StatusDurations) == 0 {
			continue
		}

		last := &taskDur.StatusDurations[len(taskDur.StatusDurations)-1]
		if last.Duration == 0.0 {
			// Status je još uvek aktivan, izračunaj koliko traje do sad
			last.Duration = float64(now-last.Timestamp) / 3600.0 // u satima
		}

		durations[taskID] = taskDur
	}

	return durations
}
