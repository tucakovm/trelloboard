package handlers

import (
	"analytics-service/models"
	proto "analytics-service/proto/analytics"
	"analytics-service/services"
	"context"
	code "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type AnalyticsHandler struct {
	service service.AnalyticsService
	proto.UnimplementedAnalyticsServiceServer
	Tracer trace.Tracer
}

// NewAnalyticsHandler initializes a new AnalyticsHandler
func NewAnalyticsHandler(service service.AnalyticsService, tracer trace.Tracer) (AnalyticsHandler, error) {
	return AnalyticsHandler{
		service: service,
		Tracer:  tracer,
	}, nil
}

// GetAnalytics fetches analytics data for a specific project
func (h AnalyticsHandler) GetAllByProjectId(ctx context.Context, req *proto.GetAnalyticsRequest) (*proto.GetAnalyticsResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.GetAnalytics")
	defer span.End()

	// Extract the project ID from the request
	projectID := req.ProjectId

	// Fetch analytics data from the service layer
	analytics, err := h.service.GetAnalytics(ctx, projectID)
	if err != nil {
		span.SetStatus(code.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "Failed to retrieve analytics.")
	}

	response := &proto.GetAnalyticsResponse{
		Analytic: &proto.Analytic{
			ProjectId:           analytics.ProjectID,
			TotalTasks:          analytics.TotalTasks,
			StatusCounts:        convertStatusCountsToProto(analytics.StatusCounts),
			TaskStatusDurations: convertTaskDurationsToProto(analytics.TaskStatusDurations),
			MemberTasks:         convertMemberTasksToProto(analytics.MemberTasks),
			FinishedEarly:       analytics.FinishedEarly,
		},
	}

	log.Println("analitike:", analytics)

	return response, nil
}

func convertStatusCountsToProto(statusCounts map[string]int32) map[string]int32 {
	result := make(map[string]int32)
	for k, v := range statusCounts {
		result[k] = v
	}
	return result
}

// Helper function to convert task durations to proto format
func convertTaskDurationsToProto(durations map[string]models.TaskDurations) map[string]*proto.TaskDurations {
	result := make(map[string]*proto.TaskDurations)
	for taskID, taskDuration := range durations {
		var protoStatusDurations []*proto.TaskStatusDuration
		// Convert each task's status durations to the proto format
		for _, statusDuration := range taskDuration.StatusDurations {
			protoStatusDurations = append(protoStatusDurations, &proto.TaskStatusDuration{
				Status:   statusDuration.Status,
				Duration: float32(statusDuration.Duration),
			})
		}
		result[taskID] = &proto.TaskDurations{
			TaskId:          taskID,
			StatusDurations: protoStatusDurations,
		}
	}
	return result
}

// Helper function to convert member tasks to proto format
func convertMemberTasksToProto(memberTasks map[string]models.TaskAssignments) map[string]*proto.MemberTasks {
	result := make(map[string]*proto.MemberTasks)
	for memberID, member := range memberTasks {
		result[memberID] = &proto.MemberTasks{
			MemberId: memberID,
			Tasks:    member.Tasks,
		}
	}
	return result
}
