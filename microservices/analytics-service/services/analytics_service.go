package service

import (
	"analytics-service/models"
	proto "analytics-service/proto/analytics"
	"analytics-service/repositories"
	"context"
	"log"
	"time"

	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnalyticsService struct {
	repo   repositories.AnalyticsRepo // Correct repository name for analytics
	Tracer trace.Tracer
}

// NewAnalyticsService initializes a new AnalyticsService
func NewAnalyticsService(repo repositories.AnalyticsRepo, tracer trace.Tracer) *AnalyticsService {
	return &AnalyticsService{
		repo:   repo,
		Tracer: tracer,
	}
}

// Create creates a new analytics entry in the database
func (s *AnalyticsService) Create(ctx context.Context, req *proto.Analytic) error {
	ctx, span := s.Tracer.Start(ctx, "s.CreateAnalytics")
	defer span.End()

	// Convert proto.Analytic to models.Analytic
	newAnalytic := &models.Analytic{
		ProjectID:           req.ProjectId,
		TotalTasks:          req.TotalTasks,
		StatusCounts:        req.StatusCounts,
		TaskStatusDurations: convertProtoTaskDurationsToModel(req.TaskStatusDurations),
		MemberTasks:         convertProtoMemberTasksToModel(req.MemberTasks),
		FinishedEarly:       req.FinishedEarly,
	}

	log.Printf("Creating analytics entry: %+v\n", newAnalytic)

	// Insert analytics entry into the repository
	if err := s.repo.InsertAnalytics(ctx, newAnalytic); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to create analytics entry in the database.")
	}

	return nil
}

// GetAnalytics fetches analytics data for a specific project
func (s *AnalyticsService) GetAnalytics(ctx context.Context, projectID string) (*models.Analytic, error) {
	ctx, span := s.Tracer.Start(ctx, "s.GetAnalytics")
	defer span.End()

	// Fetch analytics data from the repository
	analytics, err := s.repo.GetAnalyticsByProject(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to fetch analytics from the database.")
	}

	return analytics, nil
}

// Helper function to convert proto.TaskDurations to models.TaskStatusDurations
func convertProtoTaskDurationsToModel(protoDurations map[string]*proto.TaskDurations) map[string]models.TaskDurations {
	modelDurations := make(map[string]models.TaskDurations)
	for taskID, protoDuration := range protoDurations {
		var taskStatusDurations []models.TaskStatusDuration
		for _, protoStatusDuration := range protoDuration.StatusDurations {
			taskStatusDurations = append(taskStatusDurations, models.TaskStatusDuration{
				Status:   protoStatusDuration.Status,
				Duration: float64(protoStatusDuration.Duration),
			})
		}
		modelDurations[taskID] = models.TaskDurations{
			StatusDurations: taskStatusDurations,
		}
	}
	return modelDurations
}

// Helper function to convert proto.MemberTasks to models.MemberTasks
func convertProtoMemberTasksToModel(protoMembers map[string]*proto.MemberTasks) map[string]models.TaskAssignments {
	modelMembers := make(map[string]models.TaskAssignments)
	for memberID, protoMember := range protoMembers {
		modelMembers[memberID] = models.TaskAssignments{
			Tasks: protoMember.Tasks,
		}
	}
	return modelMembers
}

//Methods for nats

func (s *AnalyticsService) UpdateTaskCount(ctx context.Context, projectID string, taskID string, countDelta int) error {
	ctx, span := s.Tracer.Start(ctx, "s.UpdateTaskCount")
	defer span.End()

	analytics, err := s.ensureAnalyticsExists(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to ensure analytics for project.")
	}

	analytics.TotalTasks += int32(countDelta)

	if countDelta > 0 {
		analytics.StatusCounts["Pending"]++

		analytics.TaskStatusDurations[taskID] = models.TaskDurations{
			TaskID: taskID,
			StatusDurations: []models.TaskStatusDuration{
				{
					Status:    "Pending",
					Duration:  0.0,
					Timestamp: time.Now().Unix(),
				},
			},
		}
	}

	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update analytics.")
	}

	return nil
}

func (s *AnalyticsService) UpdateTaskStatus(ctx context.Context, projectID, taskID, newStatus string) error {
	ctx, span := s.Tracer.Start(ctx, "s.UpdateTaskStatus")
	defer span.End()

	analytics, err := s.ensureAnalyticsExists(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to ensure analytics for project.")
	}

	now := time.Now().Unix()
	taskDurations := analytics.TaskStatusDurations[taskID]

	if len(taskDurations.StatusDurations) == 0 {
		analytics.StatusCounts[newStatus]++
	} else {
		// poslednji status
		lastIdx := len(taskDurations.StatusDurations) - 1
		prev := &taskDurations.StatusDurations[lastIdx]

		if prev.Status != newStatus {
			// Izračunaj trajanje prethodnog statusa
			duration := float64(now - prev.Timestamp)
			prev.Duration += duration / 3600.0 // pretvori u sate

			// Ažuriraj status count
			analytics.StatusCounts[prev.Status]--
			analytics.StatusCounts[newStatus]++
		} else {
			// isti status, ništa ne menjamo
			return nil
		}
	}

	// Dodaj novi status sa vremenom početka
	taskDurations.StatusDurations = append(taskDurations.StatusDurations, models.TaskStatusDuration{
		Status:    newStatus,
		Duration:  0.0,
		Timestamp: now,
	})

	analytics.TaskStatusDurations[taskID] = taskDurations

	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update task status in analytics.")
	}

	return nil
}

func (s *AnalyticsService) AddMemberToTask(ctx context.Context, projectID, taskID, username string) error {
	ctx, span := s.Tracer.Start(ctx, "s.AddMemberToTask")
	defer span.End()

	analytics, err := s.ensureAnalyticsExists(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to ensure analytics for project.")
	}

	if analytics.MemberTasks == nil {
		analytics.MemberTasks = make(map[string]models.TaskAssignments)
	}

	memberAssignments := analytics.MemberTasks[username]

	for _, existingTask := range memberAssignments.Tasks {
		if existingTask == taskID {
			return nil // već postoji, ne dodaj ponovo
		}
	}

	memberAssignments.Tasks = append(memberAssignments.Tasks, taskID)
	analytics.MemberTasks[username] = memberAssignments

	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update member assignments in analytics.")
	}

	return nil
}

func (s *AnalyticsService) RemoveMemberFromTask(ctx context.Context, projectID, taskID, username string) error {
	ctx, span := s.Tracer.Start(ctx, "s.RemoveMemberFromTask")
	defer span.End()

	analytics, err := s.ensureAnalyticsExists(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to ensure analytics for project.")
	}

	if analytics.MemberTasks != nil {
		memberTasks, ok := analytics.MemberTasks[username]
		if ok {
			// Remove taskID from member's tasks
			for i, task := range memberTasks.Tasks {
				if task == taskID {
					memberTasks.Tasks = append(memberTasks.Tasks[:i], memberTasks.Tasks[i+1:]...)
					break
				}
			}

			// If no more tasks, remove the member entirely
			if len(memberTasks.Tasks) == 0 {
				delete(analytics.MemberTasks, username)
			} else {
				analytics.MemberTasks[username] = memberTasks
			}
		}
	}

	// Update analytics
	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update member assignments in analytics.")
	}

	return nil
}

func (s *AnalyticsService) ensureAnalyticsExists(ctx context.Context, projectID string) (*models.Analytic, error) {
	analytics, err := s.repo.GetAnalyticsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if analytics == nil {
		analytics = &models.Analytic{
			ProjectID:           projectID,
			TotalTasks:          0,
			StatusCounts:        map[string]int32{"Pending": 0, "Working": 0, "Done": 0},
			TaskStatusDurations: make(map[string]models.TaskDurations),
			MemberTasks:         make(map[string]models.TaskAssignments),
			FinishedEarly:       false, // ✅ bool vrednost umesto []string
		}

		if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
			return nil, err
		}
	}

	return analytics, nil
}
