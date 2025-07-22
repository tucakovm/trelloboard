package handlers

import (
	nats_helper "api-composer/nats_helper"
	proto "api-composer/proto/composer"
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type ComposerHandler struct {
	proto.UnimplementedApiComposerServer
	Tracer   trace.Tracer
	natsConn *nats.Conn
}

func NewConnectionHandler(Tracer trace.Tracer, natsConn *nats.Conn) (ComposerHandler, error) {
	return ComposerHandler{
		Tracer:   Tracer,
		natsConn: natsConn,
	}, nil
}

func (h ComposerHandler) Get(ctx context.Context, req *proto.GetReq) (*proto.GetRes, error) {
	ctx, span := h.Tracer.Start(ctx, "h.apiComposition")
	defer span.End()

	headers := nats.Header{}
	headers.Set(nats_helper.TRACE_ID, span.SpanContext().TraceID().String())
	headers.Set(nats_helper.SPAN_ID, span.SpanContext().SpanID().String())

	subjectWorkflow := "get-workflow-apiComp"
	subjectTasks := "get-tasks-apiComp"

	replySubjectWorkflow := nats.NewInbox()
	replySubjectTasks := nats.NewInbox()

	message := map[string]string{"ProjectId": req.ProjectId}
	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal message: %v", err)
		return nil, status.Error(codes.Internal, "Failed to marshal message")
	}

	replyChanWorkflow := make(chan *nats.Msg, 1)
	replyChanTasks := make(chan *nats.Msg, 1)

	defer close(replyChanWorkflow)
	defer close(replyChanTasks)

	subWorkflow, err := h.natsConn.ChanSubscribe(replySubjectWorkflow, replyChanWorkflow)
	if err != nil {
		log.Printf("[ERROR] Failed to subscribe to workflow subject: %v", err)
		return nil, status.Error(codes.Internal, "Failed to subscribe to workflow subject")
	}
	defer subWorkflow.Unsubscribe()

	subTasks, err := h.natsConn.ChanSubscribe(replySubjectTasks, replyChanTasks)
	if err != nil {
		log.Printf("[ERROR] Failed to subscribe to tasks subject: %v", err)
		return nil, status.Error(codes.Internal, "Failed to subscribe to tasks subject")
	}
	defer subTasks.Unsubscribe()

	err = h.natsConn.PublishMsg(&nats.Msg{
		Subject: subjectWorkflow,
		Reply:   replySubjectWorkflow,
		Data:    messageData,
		Header:  headers,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to publish workflow request: %v", err)
		return nil, status.Error(codes.Internal, "Failed to publish workflow request")
	}

	err = h.natsConn.PublishMsg(&nats.Msg{
		Subject: subjectTasks,
		Reply:   replySubjectTasks,
		Data:    messageData,
		Header:  headers,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to publish tasks request: %v", err)
		return nil, status.Error(codes.Internal, "Failed to publish tasks request")
	}

	var workflow *proto.Workflow
	var tasks []*proto.Task
	timeout := 5 * time.Second

	for i := 0; i < 2; i++ {
		select {
		case msg := <-replyChanWorkflow:
			var workflowResponse struct {
				Workflow *proto.Workflow `json:"workflow"`
			}
			log.Println("data from workflow:", msg.Data)
			if err := json.Unmarshal(msg.Data, &workflowResponse.Workflow); err != nil {
				log.Printf("[ERROR] Failed to unmarshal workflow response: %v", err)
			} else {
				workflow = workflowResponse.Workflow
				log.Println("[DEBUG] Received workflow response", workflowResponse)
			}

		case msg := <-replyChanTasks:
			var tasksResponse struct {
				Tasks []proto.Task `json:"tasks"`
			}
			if err := json.Unmarshal(msg.Data, &tasksResponse.Tasks); err != nil {
				log.Printf("[ERROR] Failed to unmarshal tasks response: %v", err)
			} else {
				tasks = make([]*proto.Task, len(tasksResponse.Tasks))
				for i := range tasksResponse.Tasks {
					tasks[i] = &tasksResponse.Tasks[i]
				}
				log.Println("[DEBUG] Received tasks response", tasksResponse)
			}

		case <-time.After(timeout):
			log.Println("[ERROR] Timeout waiting for workflow/tasks response")
			return nil, status.Error(codes.DeadlineExceeded, "Timeout waiting for workflow/tasks response")
		}
	}

	res := &proto.ApiCompositionObject{
		Tasks:    tasks,
		Workflow: workflow,
	}
	response := &proto.GetRes{
		Aco: res,
	}

	return response, nil
}
