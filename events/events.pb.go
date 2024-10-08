// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.2
// source: events.proto

package events

import (
	tes "github.com/ohsu-comp-bio/funnel/tes"
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

type Type int32

const (
	Type_UNKNOWN             Type = 0
	Type_TASK_STATE          Type = 1
	Type_TASK_START_TIME     Type = 2
	Type_TASK_END_TIME       Type = 3
	Type_TASK_OUTPUTS        Type = 4
	Type_TASK_METADATA       Type = 5
	Type_EXECUTOR_START_TIME Type = 6
	Type_EXECUTOR_END_TIME   Type = 7
	Type_EXECUTOR_EXIT_CODE  Type = 8
	Type_EXECUTOR_STDOUT     Type = 11
	Type_EXECUTOR_STDERR     Type = 12
	Type_SYSTEM_LOG          Type = 13
	Type_TASK_CREATED        Type = 14
)

// Enum value maps for Type.
var (
	Type_name = map[int32]string{
		0:  "UNKNOWN",
		1:  "TASK_STATE",
		2:  "TASK_START_TIME",
		3:  "TASK_END_TIME",
		4:  "TASK_OUTPUTS",
		5:  "TASK_METADATA",
		6:  "EXECUTOR_START_TIME",
		7:  "EXECUTOR_END_TIME",
		8:  "EXECUTOR_EXIT_CODE",
		11: "EXECUTOR_STDOUT",
		12: "EXECUTOR_STDERR",
		13: "SYSTEM_LOG",
		14: "TASK_CREATED",
	}
	Type_value = map[string]int32{
		"UNKNOWN":             0,
		"TASK_STATE":          1,
		"TASK_START_TIME":     2,
		"TASK_END_TIME":       3,
		"TASK_OUTPUTS":        4,
		"TASK_METADATA":       5,
		"EXECUTOR_START_TIME": 6,
		"EXECUTOR_END_TIME":   7,
		"EXECUTOR_EXIT_CODE":  8,
		"EXECUTOR_STDOUT":     11,
		"EXECUTOR_STDERR":     12,
		"SYSTEM_LOG":          13,
		"TASK_CREATED":        14,
	}
)

func (x Type) Enum() *Type {
	p := new(Type)
	*p = x
	return p
}

func (x Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Type) Descriptor() protoreflect.EnumDescriptor {
	return file_events_proto_enumTypes[0].Descriptor()
}

func (Type) Type() protoreflect.EnumType {
	return &file_events_proto_enumTypes[0]
}

func (x Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Type.Descriptor instead.
func (Type) EnumDescriptor() ([]byte, []int) {
	return file_events_proto_rawDescGZIP(), []int{0}
}

type Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value map[string]string `protobuf:"bytes,1,rep,name=value,proto3" json:"value,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_events_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

func (x *Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_events_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metadata.ProtoReflect.Descriptor instead.
func (*Metadata) Descriptor() ([]byte, []int) {
	return file_events_proto_rawDescGZIP(), []int{0}
}

func (x *Metadata) GetValue() map[string]string {
	if x != nil {
		return x.Value
	}
	return nil
}

type Outputs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value []*tes.OutputFileLog `protobuf:"bytes,1,rep,name=value,proto3" json:"value,omitempty"`
}

func (x *Outputs) Reset() {
	*x = Outputs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_events_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Outputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Outputs) ProtoMessage() {}

func (x *Outputs) ProtoReflect() protoreflect.Message {
	mi := &file_events_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Outputs.ProtoReflect.Descriptor instead.
func (*Outputs) Descriptor() ([]byte, []int) {
	return file_events_proto_rawDescGZIP(), []int{1}
}

func (x *Outputs) GetValue() []*tes.OutputFileLog {
	if x != nil {
		return x.Value
	}
	return nil
}

type SystemLog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg    string            `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	Level  string            `protobuf:"bytes,2,opt,name=level,proto3" json:"level,omitempty"`
	Fields map[string]string `protobuf:"bytes,3,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *SystemLog) Reset() {
	*x = SystemLog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_events_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemLog) ProtoMessage() {}

func (x *SystemLog) ProtoReflect() protoreflect.Message {
	mi := &file_events_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemLog.ProtoReflect.Descriptor instead.
func (*SystemLog) Descriptor() ([]byte, []int) {
	return file_events_proto_rawDescGZIP(), []int{2}
}

func (x *SystemLog) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *SystemLog) GetLevel() string {
	if x != nil {
		return x.Level
	}
	return ""
}

func (x *SystemLog) GetFields() map[string]string {
	if x != nil {
		return x.Fields
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Timestamp string `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Types that are assignable to Data:
	//
	//	*Event_State
	//	*Event_StartTime
	//	*Event_EndTime
	//	*Event_Outputs
	//	*Event_Metadata
	//	*Event_ExitCode
	//	*Event_Stdout
	//	*Event_Stderr
	//	*Event_SystemLog
	//	*Event_Task
	Data    isEvent_Data `protobuf_oneof:"data"`
	Attempt uint32       `protobuf:"varint,16,opt,name=attempt,proto3" json:"attempt,omitempty"`
	Index   uint32       `protobuf:"varint,17,opt,name=index,proto3" json:"index,omitempty"`
	Type    Type         `protobuf:"varint,18,opt,name=type,proto3,enum=events.Type" json:"type,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_events_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_events_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_events_proto_rawDescGZIP(), []int{3}
}

func (x *Event) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Event) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

func (m *Event) GetData() isEvent_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *Event) GetState() tes.State {
	if x, ok := x.GetData().(*Event_State); ok {
		return x.State
	}
	return tes.State(0)
}

func (x *Event) GetStartTime() string {
	if x, ok := x.GetData().(*Event_StartTime); ok {
		return x.StartTime
	}
	return ""
}

func (x *Event) GetEndTime() string {
	if x, ok := x.GetData().(*Event_EndTime); ok {
		return x.EndTime
	}
	return ""
}

func (x *Event) GetOutputs() *Outputs {
	if x, ok := x.GetData().(*Event_Outputs); ok {
		return x.Outputs
	}
	return nil
}

func (x *Event) GetMetadata() *Metadata {
	if x, ok := x.GetData().(*Event_Metadata); ok {
		return x.Metadata
	}
	return nil
}

func (x *Event) GetExitCode() int32 {
	if x, ok := x.GetData().(*Event_ExitCode); ok {
		return x.ExitCode
	}
	return 0
}

func (x *Event) GetStdout() string {
	if x, ok := x.GetData().(*Event_Stdout); ok {
		return x.Stdout
	}
	return ""
}

func (x *Event) GetStderr() string {
	if x, ok := x.GetData().(*Event_Stderr); ok {
		return x.Stderr
	}
	return ""
}

func (x *Event) GetSystemLog() *SystemLog {
	if x, ok := x.GetData().(*Event_SystemLog); ok {
		return x.SystemLog
	}
	return nil
}

func (x *Event) GetTask() *tes.Task {
	if x, ok := x.GetData().(*Event_Task); ok {
		return x.Task
	}
	return nil
}

func (x *Event) GetAttempt() uint32 {
	if x != nil {
		return x.Attempt
	}
	return 0
}

func (x *Event) GetIndex() uint32 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *Event) GetType() Type {
	if x != nil {
		return x.Type
	}
	return Type_UNKNOWN
}

type isEvent_Data interface {
	isEvent_Data()
}

type Event_State struct {
	State tes.State `protobuf:"varint,3,opt,name=state,proto3,enum=tes.State,oneof"`
}

type Event_StartTime struct {
	StartTime string `protobuf:"bytes,4,opt,name=start_time,json=startTime,proto3,oneof"`
}

type Event_EndTime struct {
	EndTime string `protobuf:"bytes,5,opt,name=end_time,json=endTime,proto3,oneof"`
}

type Event_Outputs struct {
	Outputs *Outputs `protobuf:"bytes,6,opt,name=outputs,proto3,oneof"`
}

type Event_Metadata struct {
	Metadata *Metadata `protobuf:"bytes,7,opt,name=metadata,proto3,oneof"`
}

type Event_ExitCode struct {
	ExitCode int32 `protobuf:"varint,10,opt,name=exit_code,json=exitCode,proto3,oneof"`
}

type Event_Stdout struct {
	Stdout string `protobuf:"bytes,13,opt,name=stdout,proto3,oneof"`
}

type Event_Stderr struct {
	Stderr string `protobuf:"bytes,14,opt,name=stderr,proto3,oneof"`
}

type Event_SystemLog struct {
	SystemLog *SystemLog `protobuf:"bytes,15,opt,name=system_log,json=systemLog,proto3,oneof"`
}

type Event_Task struct {
	Task *tes.Task `protobuf:"bytes,19,opt,name=task,proto3,oneof"`
}

func (*Event_State) isEvent_Data() {}

func (*Event_StartTime) isEvent_Data() {}

func (*Event_EndTime) isEvent_Data() {}

func (*Event_Outputs) isEvent_Data() {}

func (*Event_Metadata) isEvent_Data() {}

func (*Event_ExitCode) isEvent_Data() {}

func (*Event_Stdout) isEvent_Data() {}

func (*Event_Stderr) isEvent_Data() {}

func (*Event_SystemLog) isEvent_Data() {}

func (*Event_Task) isEvent_Data() {}

type WriteEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WriteEventResponse) Reset() {
	*x = WriteEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_events_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteEventResponse) ProtoMessage() {}

func (x *WriteEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_events_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteEventResponse.ProtoReflect.Descriptor instead.
func (*WriteEventResponse) Descriptor() ([]byte, []int) {
	return file_events_proto_rawDescGZIP(), []int{4}
}

var File_events_proto protoreflect.FileDescriptor

var file_events_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x1a, 0x09, 0x74, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x77, 0x0a, 0x08, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x31, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x1a, 0x38, 0x0a, 0x0a, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x33, 0x0a, 0x07, 0x4f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x73, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x74, 0x65, 0x73, 0x2e, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x46, 0x69, 0x6c, 0x65, 0x4c, 0x6f, 0x67, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22,
	0xa5, 0x01, 0x0a, 0x09, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4c, 0x6f, 0x67, 0x12, 0x10, 0x0a,
	0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12,
	0x14, 0x0a, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x35, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x4c, 0x6f, 0x67, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x1a, 0x39, 0x0a, 0x0b,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xf6, 0x03, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12,
	0x22, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0a,
	0x2e, 0x74, 0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x48, 0x00, 0x52, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x12, 0x1f, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x1b, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x65, 0x6e, 0x64, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x2b, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x73, 0x48, 0x00, 0x52, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x12, 0x2e,
	0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x48, 0x00, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1d,
	0x0a, 0x09, 0x65, 0x78, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x05, 0x48, 0x00, 0x52, 0x08, 0x65, 0x78, 0x69, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a,
	0x06, 0x73, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x06, 0x73, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x12, 0x18, 0x0a, 0x06, 0x73, 0x74, 0x64, 0x65, 0x72,
	0x72, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x73, 0x74, 0x64, 0x65, 0x72,
	0x72, 0x12, 0x32, 0x0a, 0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f, 0x6c, 0x6f, 0x67, 0x18,
	0x0f, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x4c, 0x6f, 0x67, 0x48, 0x00, 0x52, 0x09, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x4c, 0x6f, 0x67, 0x12, 0x1f, 0x0a, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x18, 0x13, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x74, 0x65, 0x73, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x48, 0x00,
	0x52, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x74, 0x74, 0x65, 0x6d, 0x70,
	0x74, 0x18, 0x10, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x61, 0x74, 0x74, 0x65, 0x6d, 0x70, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x20, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x12,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x22, 0x14, 0x0a, 0x12, 0x57, 0x72, 0x69, 0x74, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2a, 0x84, 0x02, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a,
	0x54, 0x41, 0x53, 0x4b, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f,
	0x54, 0x41, 0x53, 0x4b, 0x5f, 0x53, 0x54, 0x41, 0x52, 0x54, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x10,
	0x02, 0x12, 0x11, 0x0a, 0x0d, 0x54, 0x41, 0x53, 0x4b, 0x5f, 0x45, 0x4e, 0x44, 0x5f, 0x54, 0x49,
	0x4d, 0x45, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x41, 0x53, 0x4b, 0x5f, 0x4f, 0x55, 0x54,
	0x50, 0x55, 0x54, 0x53, 0x10, 0x04, 0x12, 0x11, 0x0a, 0x0d, 0x54, 0x41, 0x53, 0x4b, 0x5f, 0x4d,
	0x45, 0x54, 0x41, 0x44, 0x41, 0x54, 0x41, 0x10, 0x05, 0x12, 0x17, 0x0a, 0x13, 0x45, 0x58, 0x45,
	0x43, 0x55, 0x54, 0x4f, 0x52, 0x5f, 0x53, 0x54, 0x41, 0x52, 0x54, 0x5f, 0x54, 0x49, 0x4d, 0x45,
	0x10, 0x06, 0x12, 0x15, 0x0a, 0x11, 0x45, 0x58, 0x45, 0x43, 0x55, 0x54, 0x4f, 0x52, 0x5f, 0x45,
	0x4e, 0x44, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x10, 0x07, 0x12, 0x16, 0x0a, 0x12, 0x45, 0x58, 0x45,
	0x43, 0x55, 0x54, 0x4f, 0x52, 0x5f, 0x45, 0x58, 0x49, 0x54, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x10,
	0x08, 0x12, 0x13, 0x0a, 0x0f, 0x45, 0x58, 0x45, 0x43, 0x55, 0x54, 0x4f, 0x52, 0x5f, 0x53, 0x54,
	0x44, 0x4f, 0x55, 0x54, 0x10, 0x0b, 0x12, 0x13, 0x0a, 0x0f, 0x45, 0x58, 0x45, 0x43, 0x55, 0x54,
	0x4f, 0x52, 0x5f, 0x53, 0x54, 0x44, 0x45, 0x52, 0x52, 0x10, 0x0c, 0x12, 0x0e, 0x0a, 0x0a, 0x53,
	0x59, 0x53, 0x54, 0x45, 0x4d, 0x5f, 0x4c, 0x4f, 0x47, 0x10, 0x0d, 0x12, 0x10, 0x0a, 0x0c, 0x54,
	0x41, 0x53, 0x4b, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x44, 0x10, 0x0e, 0x32, 0x49, 0x0a,
	0x0c, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x39, 0x0a,
	0x0a, 0x57, 0x72, 0x69, 0x74, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x0d, 0x2e, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x1a, 0x1a, 0x2e, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x28, 0x5a, 0x26, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x68, 0x73, 0x75, 0x2d, 0x63, 0x6f, 0x6d, 0x70,
	0x2d, 0x62, 0x69, 0x6f, 0x2f, 0x66, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2f, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_events_proto_rawDescOnce sync.Once
	file_events_proto_rawDescData = file_events_proto_rawDesc
)

func file_events_proto_rawDescGZIP() []byte {
	file_events_proto_rawDescOnce.Do(func() {
		file_events_proto_rawDescData = protoimpl.X.CompressGZIP(file_events_proto_rawDescData)
	})
	return file_events_proto_rawDescData
}

var file_events_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_events_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_events_proto_goTypes = []interface{}{
	(Type)(0),                  // 0: events.Type
	(*Metadata)(nil),           // 1: events.Metadata
	(*Outputs)(nil),            // 2: events.Outputs
	(*SystemLog)(nil),          // 3: events.SystemLog
	(*Event)(nil),              // 4: events.Event
	(*WriteEventResponse)(nil), // 5: events.WriteEventResponse
	nil,                        // 6: events.Metadata.ValueEntry
	nil,                        // 7: events.SystemLog.FieldsEntry
	(*tes.OutputFileLog)(nil),  // 8: tes.OutputFileLog
	(tes.State)(0),             // 9: tes.State
	(*tes.Task)(nil),           // 10: tes.Task
}
var file_events_proto_depIdxs = []int32{
	6,  // 0: events.Metadata.value:type_name -> events.Metadata.ValueEntry
	8,  // 1: events.Outputs.value:type_name -> tes.OutputFileLog
	7,  // 2: events.SystemLog.fields:type_name -> events.SystemLog.FieldsEntry
	9,  // 3: events.Event.state:type_name -> tes.State
	2,  // 4: events.Event.outputs:type_name -> events.Outputs
	1,  // 5: events.Event.metadata:type_name -> events.Metadata
	3,  // 6: events.Event.system_log:type_name -> events.SystemLog
	10, // 7: events.Event.task:type_name -> tes.Task
	0,  // 8: events.Event.type:type_name -> events.Type
	4,  // 9: events.EventService.WriteEvent:input_type -> events.Event
	5,  // 10: events.EventService.WriteEvent:output_type -> events.WriteEventResponse
	10, // [10:11] is the sub-list for method output_type
	9,  // [9:10] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_events_proto_init() }
func file_events_proto_init() {
	if File_events_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_events_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metadata); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_events_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Outputs); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_events_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemLog); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_events_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_events_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteEventResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_events_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*Event_State)(nil),
		(*Event_StartTime)(nil),
		(*Event_EndTime)(nil),
		(*Event_Outputs)(nil),
		(*Event_Metadata)(nil),
		(*Event_ExitCode)(nil),
		(*Event_Stdout)(nil),
		(*Event_Stderr)(nil),
		(*Event_SystemLog)(nil),
		(*Event_Task)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_events_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_events_proto_goTypes,
		DependencyIndexes: file_events_proto_depIdxs,
		EnumInfos:         file_events_proto_enumTypes,
		MessageInfos:      file_events_proto_msgTypes,
	}.Build()
	File_events_proto = out.File
	file_events_proto_rawDesc = nil
	file_events_proto_goTypes = nil
	file_events_proto_depIdxs = nil
}
