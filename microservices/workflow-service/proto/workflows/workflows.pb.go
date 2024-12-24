// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v5.29.2
// source: workflows.proto

package workflows

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Messages for Workflow Service
type VoidResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VoidResponse) Reset() {
	*x = VoidResponse{}
	mi := &file_workflows_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VoidResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VoidResponse) ProtoMessage() {}

func (x *VoidResponse) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VoidResponse.ProtoReflect.Descriptor instead.
func (*VoidResponse) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{0}
}

type CreateWorkflowReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProjectId     string                 `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	ProjectName   string                 `protobuf:"bytes,2,opt,name=project_name,json=projectName,proto3" json:"project_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateWorkflowReq) Reset() {
	*x = CreateWorkflowReq{}
	mi := &file_workflows_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateWorkflowReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateWorkflowReq) ProtoMessage() {}

func (x *CreateWorkflowReq) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateWorkflowReq.ProtoReflect.Descriptor instead.
func (*CreateWorkflowReq) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{1}
}

func (x *CreateWorkflowReq) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *CreateWorkflowReq) GetProjectName() string {
	if x != nil {
		return x.ProjectName
	}
	return ""
}

type AddTaskReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProjectId     string                 `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	Task          *Task                  `protobuf:"bytes,2,opt,name=task,proto3" json:"task,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddTaskReq) Reset() {
	*x = AddTaskReq{}
	mi := &file_workflows_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddTaskReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddTaskReq) ProtoMessage() {}

func (x *AddTaskReq) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddTaskReq.ProtoReflect.Descriptor instead.
func (*AddTaskReq) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{2}
}

func (x *AddTaskReq) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *AddTaskReq) GetTask() *Task {
	if x != nil {
		return x.Task
	}
	return nil
}

type GetWorkflowReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProjectId     string                 `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetWorkflowReq) Reset() {
	*x = GetWorkflowReq{}
	mi := &file_workflows_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetWorkflowReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetWorkflowReq) ProtoMessage() {}

func (x *GetWorkflowReq) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetWorkflowReq.ProtoReflect.Descriptor instead.
func (*GetWorkflowReq) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{3}
}

func (x *GetWorkflowReq) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

type GetWorkflowRes struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Workflow      *Workflow              `protobuf:"bytes,1,opt,name=workflow,proto3" json:"workflow,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetWorkflowRes) Reset() {
	*x = GetWorkflowRes{}
	mi := &file_workflows_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetWorkflowRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetWorkflowRes) ProtoMessage() {}

func (x *GetWorkflowRes) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetWorkflowRes.ProtoReflect.Descriptor instead.
func (*GetWorkflowRes) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{4}
}

func (x *GetWorkflowRes) GetWorkflow() *Workflow {
	if x != nil {
		return x.Workflow
	}
	return nil
}

type Workflow struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProjectId     string                 `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	ProjectName   string                 `protobuf:"bytes,2,opt,name=project_name,json=projectName,proto3" json:"project_name,omitempty"`
	Tasks         []*Task                `protobuf:"bytes,3,rep,name=tasks,proto3" json:"tasks,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Workflow) Reset() {
	*x = Workflow{}
	mi := &file_workflows_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Workflow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Workflow) ProtoMessage() {}

func (x *Workflow) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Workflow.ProtoReflect.Descriptor instead.
func (*Workflow) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{5}
}

func (x *Workflow) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *Workflow) GetProjectName() string {
	if x != nil {
		return x.ProjectName
	}
	return ""
}

func (x *Workflow) GetTasks() []*Task {
	if x != nil {
		return x.Tasks
	}
	return nil
}

type Task struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description   string                 `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Dependencies  []string               `protobuf:"bytes,4,rep,name=dependencies,proto3" json:"dependencies,omitempty"`
	Blocked       bool                   `protobuf:"varint,5,opt,name=blocked,proto3" json:"blocked,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Task) Reset() {
	*x = Task{}
	mi := &file_workflows_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{6}
}

func (x *Task) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Task) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Task) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Task) GetDependencies() []string {
	if x != nil {
		return x.Dependencies
	}
	return nil
}

func (x *Task) GetBlocked() bool {
	if x != nil {
		return x.Blocked
	}
	return false
}

type CheckTaskDependenciesReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProjectId     string                 `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	TaskId        string                 `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CheckTaskDependenciesReq) Reset() {
	*x = CheckTaskDependenciesReq{}
	mi := &file_workflows_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CheckTaskDependenciesReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckTaskDependenciesReq) ProtoMessage() {}

func (x *CheckTaskDependenciesReq) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckTaskDependenciesReq.ProtoReflect.Descriptor instead.
func (*CheckTaskDependenciesReq) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{7}
}

func (x *CheckTaskDependenciesReq) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *CheckTaskDependenciesReq) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type TaskDependenciesStatus struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	AllDependenciesMet bool                   `protobuf:"varint,1,opt,name=allDependenciesMet,proto3" json:"allDependenciesMet,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *TaskDependenciesStatus) Reset() {
	*x = TaskDependenciesStatus{}
	mi := &file_workflows_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskDependenciesStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskDependenciesStatus) ProtoMessage() {}

func (x *TaskDependenciesStatus) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskDependenciesStatus.ProtoReflect.Descriptor instead.
func (*TaskDependenciesStatus) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{8}
}

func (x *TaskDependenciesStatus) GetAllDependenciesMet() bool {
	if x != nil {
		return x.AllDependenciesMet
	}
	return false
}

type TaskExistsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskExistsRequest) Reset() {
	*x = TaskExistsRequest{}
	mi := &file_workflows_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskExistsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskExistsRequest) ProtoMessage() {}

func (x *TaskExistsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskExistsRequest.ProtoReflect.Descriptor instead.
func (*TaskExistsRequest) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{9}
}

func (x *TaskExistsRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type TaskExistsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Exists        bool                   `protobuf:"varint,1,opt,name=exists,proto3" json:"exists,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskExistsResponse) Reset() {
	*x = TaskExistsResponse{}
	mi := &file_workflows_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskExistsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskExistsResponse) ProtoMessage() {}

func (x *TaskExistsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskExistsResponse.ProtoReflect.Descriptor instead.
func (*TaskExistsResponse) Descriptor() ([]byte, []int) {
	return file_workflows_proto_rawDescGZIP(), []int{10}
}

func (x *TaskExistsResponse) GetExists() bool {
	if x != nil {
		return x.Exists
	}
	return false
}

var File_workflows_proto protoreflect.FileDescriptor

var file_workflows_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x0e, 0x0a, 0x0c, 0x56, 0x6f, 0x69, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x55, 0x0a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x66,
	0x6c, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x46, 0x0a, 0x0a, 0x41, 0x64, 0x64, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x04, 0x74, 0x61, 0x73, 0x6b,
	0x22, 0x2f, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52,
	0x65, 0x71, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49,
	0x64, 0x22, 0x37, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x52, 0x65, 0x73, 0x12, 0x25, 0x0a, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x52, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x22, 0x69, 0x0a, 0x08, 0x57, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1b, 0x0a, 0x05, 0x74, 0x61, 0x73, 0x6b,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x05,
	0x74, 0x61, 0x73, 0x6b, 0x73, 0x22, 0x8a, 0x01, 0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e,
	0x63, 0x69, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x64, 0x65, 0x70, 0x65,
	0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x65, 0x64, 0x22, 0x52, 0x0a, 0x18, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x54, 0x61, 0x73, 0x6b, 0x44,
	0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x12, 0x1d,
	0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a,
	0x07, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x22, 0x48, 0x0a, 0x16, 0x54, 0x61, 0x73, 0x6b, 0x44, 0x65,
	0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x2e, 0x0a, 0x12, 0x61, 0x6c, 0x6c, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63,
	0x69, 0x65, 0x73, 0x4d, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x61, 0x6c,
	0x6c, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x4d, 0x65, 0x74,
	0x22, 0x2c, 0x0a, 0x11, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x22, 0x2c,
	0x0a, 0x12, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x32, 0xf6, 0x02, 0x0a,
	0x0f, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x35, 0x0a, 0x0e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x12, 0x12, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x66,
	0x6c, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x1a, 0x0d, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x27, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x54, 0x61,
	0x73, 0x6b, 0x12, 0x0b, 0x2e, 0x41, 0x64, 0x64, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x1a,
	0x0d, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x3c, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x42,
	0x79, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x12, 0x0f, 0x2e, 0x47, 0x65, 0x74,
	0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x1a, 0x0f, 0x2e, 0x47, 0x65,
	0x74, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x3d,
	0x0a, 0x19, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x42, 0x79, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x12, 0x0f, 0x2e, 0x47, 0x65,
	0x74, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x1a, 0x0d, 0x2e, 0x56,
	0x6f, 0x69, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4d, 0x0a,
	0x15, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x54, 0x61, 0x73, 0x6b, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64,
	0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x12, 0x19, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x54, 0x61,
	0x73, 0x6b, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x1a, 0x17, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e,
	0x63, 0x69, 0x65, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x37, 0x0a, 0x0a,
	0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x12, 0x12, 0x2e, 0x54, 0x61, 0x73,
	0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13,
	0x2e, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x11, 0x5a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_workflows_proto_rawDescOnce sync.Once
	file_workflows_proto_rawDescData = file_workflows_proto_rawDesc
)

func file_workflows_proto_rawDescGZIP() []byte {
	file_workflows_proto_rawDescOnce.Do(func() {
		file_workflows_proto_rawDescData = protoimpl.X.CompressGZIP(file_workflows_proto_rawDescData)
	})
	return file_workflows_proto_rawDescData
}

var file_workflows_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_workflows_proto_goTypes = []any{
	(*VoidResponse)(nil),             // 0: VoidResponse
	(*CreateWorkflowReq)(nil),        // 1: CreateWorkflowReq
	(*AddTaskReq)(nil),               // 2: AddTaskReq
	(*GetWorkflowReq)(nil),           // 3: GetWorkflowReq
	(*GetWorkflowRes)(nil),           // 4: GetWorkflowRes
	(*Workflow)(nil),                 // 5: Workflow
	(*Task)(nil),                     // 6: Task
	(*CheckTaskDependenciesReq)(nil), // 7: CheckTaskDependenciesReq
	(*TaskDependenciesStatus)(nil),   // 8: TaskDependenciesStatus
	(*TaskExistsRequest)(nil),        // 9: TaskExistsRequest
	(*TaskExistsResponse)(nil),       // 10: TaskExistsResponse
}
var file_workflows_proto_depIdxs = []int32{
	6,  // 0: AddTaskReq.task:type_name -> Task
	5,  // 1: GetWorkflowRes.workflow:type_name -> Workflow
	6,  // 2: Workflow.tasks:type_name -> Task
	1,  // 3: WorkflowService.CreateWorkflow:input_type -> CreateWorkflowReq
	2,  // 4: WorkflowService.AddTask:input_type -> AddTaskReq
	3,  // 5: WorkflowService.GetWorkflowByProjectID:input_type -> GetWorkflowReq
	3,  // 6: WorkflowService.DeleteWorkflowByProjectID:input_type -> GetWorkflowReq
	7,  // 7: WorkflowService.CheckTaskDependencies:input_type -> CheckTaskDependenciesReq
	9,  // 8: WorkflowService.TaskExists:input_type -> TaskExistsRequest
	0,  // 9: WorkflowService.CreateWorkflow:output_type -> VoidResponse
	0,  // 10: WorkflowService.AddTask:output_type -> VoidResponse
	4,  // 11: WorkflowService.GetWorkflowByProjectID:output_type -> GetWorkflowRes
	0,  // 12: WorkflowService.DeleteWorkflowByProjectID:output_type -> VoidResponse
	8,  // 13: WorkflowService.CheckTaskDependencies:output_type -> TaskDependenciesStatus
	10, // 14: WorkflowService.TaskExists:output_type -> TaskExistsResponse
	9,  // [9:15] is the sub-list for method output_type
	3,  // [3:9] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_workflows_proto_init() }
func file_workflows_proto_init() {
	if File_workflows_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_workflows_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_workflows_proto_goTypes,
		DependencyIndexes: file_workflows_proto_depIdxs,
		MessageInfos:      file_workflows_proto_msgTypes,
	}.Build()
	File_workflows_proto = out.File
	file_workflows_proto_rawDesc = nil
	file_workflows_proto_goTypes = nil
	file_workflows_proto_depIdxs = nil
}
