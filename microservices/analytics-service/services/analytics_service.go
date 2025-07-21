package service

import (
	"analytics-service/models"
	proto "analytics-service/proto/analytics"
	"analytics-service/repositories"
	"context"
	"log"

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

// UpdateTaskCount updates the total task count for a project
func (s *AnalyticsService) UpdateTaskCount(ctx context.Context, projectID string, countDelta int) error {
	ctx, span := s.Tracer.Start(ctx, "s.UpdateTaskCount")
	defer span.End()

	// Fetch the current analytics
	analytics, err := s.repo.GetAnalyticsByProject(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to fetch analytics for project.")
	}

	// Update the task count
	analytics.TotalTasks += int32(countDelta)

	// Save the updated analytics back to the repository
	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update task count in analytics.")
	}

	return nil
}

// UpdateTaskStatus updates the status of a task and adjusts analytics accordingly
func (s *AnalyticsService) UpdateTaskStatus(ctx context.Context, projectID, taskID, newStatus string) error {
	ctx, span := s.Tracer.Start(ctx, "s.UpdateTaskStatus")
	defer span.End()

	// Fetch the current analytics
	analytics, err := s.repo.GetAnalyticsByProject(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to fetch analytics for project.")
	}

	// Update status counts
	previousStatus := "" // Assume logic to fetch the previous status of the task
	if previousStatus != "" {
		analytics.StatusCounts[previousStatus]--
	}
	analytics.StatusCounts[newStatus]++

	// Save the updated analytics back to the repository
	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update task status in analytics.")
	}

	return nil
}

// AddMemberToTask assigns a member to a task and updates analytics
func (s *AnalyticsService) AddMemberToTask(ctx context.Context, projectID, taskID, memberID string) error {
	ctx, span := s.Tracer.Start(ctx, "s.AddMemberToTask")
	defer span.End()

	// Fetch the current analytics
	analytics, err := s.repo.GetAnalyticsByProject(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to fetch analytics for project.")
	}

	// Add the task to the member's assignments
	if analytics.MemberTasks == nil {
		analytics.MemberTasks = make(map[string]models.TaskAssignments)
	}
	memberTasks := analytics.MemberTasks[memberID]
	memberTasks.Tasks = append(memberTasks.Tasks, taskID)
	analytics.MemberTasks[memberID] = memberTasks

	// Save the updated analytics back to the repository
	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update member assignments in analytics.")
	}

	return nil
}

// RemoveMemberFromTask removes a member from a task and updates analytics
func (s *AnalyticsService) RemoveMemberFromTask(ctx context.Context, projectID, taskID, memberID string) error {
	ctx, span := s.Tracer.Start(ctx, "s.RemoveMemberFromTask")
	defer span.End()

	// Fetch the current analytics
	analytics, err := s.repo.GetAnalyticsByProject(ctx, projectID)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to fetch analytics for project.")
	}

	// Remove the task from the member's assignments
	if analytics.MemberTasks != nil {
		memberTasks := analytics.MemberTasks[memberID]
		for i, task := range memberTasks.Tasks {
			if task == taskID {
				memberTasks.Tasks = append(memberTasks.Tasks[:i], memberTasks.Tasks[i+1:]...)
				break
			}
		}
		analytics.MemberTasks[memberID] = memberTasks
	}

	// Save the updated analytics back to the repository
	if err := s.repo.UpdateAnalytics(ctx, projectID, analytics); err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update member assignments in analytics.")
	}

	return nil
}
