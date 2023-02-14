// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: proto/server_pr.proto

package __

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

type Metrics_MType int32

const (
	Metrics_GAUGE   Metrics_MType = 0
	Metrics_COUNTER Metrics_MType = 1
)

// Enum value maps for Metrics_MType.
var (
	Metrics_MType_name = map[int32]string{
		0: "GAUGE",
		1: "COUNTER",
	}
	Metrics_MType_value = map[string]int32{
		"GAUGE":   0,
		"COUNTER": 1,
	}
)

func (x Metrics_MType) Enum() *Metrics_MType {
	p := new(Metrics_MType)
	*p = x
	return p
}

func (x Metrics_MType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Metrics_MType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_server_pr_proto_enumTypes[0].Descriptor()
}

func (Metrics_MType) Type() protoreflect.EnumType {
	return &file_proto_server_pr_proto_enumTypes[0]
}

func (x Metrics_MType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Metrics_MType.Descriptor instead.
func (Metrics_MType) EnumDescriptor() ([]byte, []int) {
	return file_proto_server_pr_proto_rawDescGZIP(), []int{0, 0}
}

type Metrics struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string        `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Mtype Metrics_MType `protobuf:"varint,2,opt,name=mtype,proto3,enum=main.Metrics_MType" json:"mtype,omitempty"`
	Delta int64         `protobuf:"varint,3,opt,name=delta,proto3" json:"delta,omitempty"`
	Value float64       `protobuf:"fixed64,4,opt,name=value,proto3" json:"value,omitempty"`
	Hash  string        `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *Metrics) Reset() {
	*x = Metrics{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_server_pr_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metrics) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metrics) ProtoMessage() {}

func (x *Metrics) ProtoReflect() protoreflect.Message {
	mi := &file_proto_server_pr_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metrics.ProtoReflect.Descriptor instead.
func (*Metrics) Descriptor() ([]byte, []int) {
	return file_proto_server_pr_proto_rawDescGZIP(), []int{0}
}

func (x *Metrics) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Metrics) GetMtype() Metrics_MType {
	if x != nil {
		return x.Mtype
	}
	return Metrics_GAUGE
}

func (x *Metrics) GetDelta() int64 {
	if x != nil {
		return x.Delta
	}
	return 0
}

func (x *Metrics) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *Metrics) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metrics *Metrics `protobuf:"bytes,1,opt,name=metrics,proto3" json:"metrics,omitempty"`
}

func (x *UpdateRequest) Reset() {
	*x = UpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_server_pr_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRequest) ProtoMessage() {}

func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_server_pr_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRequest.ProtoReflect.Descriptor instead.
func (*UpdateRequest) Descriptor() ([]byte, []int) {
	return file_proto_server_pr_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateRequest) GetMetrics() *Metrics {
	if x != nil {
		return x.Metrics
	}
	return nil
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"` // ошибка
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_server_pr_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_server_pr_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateResponse.ProtoReflect.Descriptor instead.
func (*UpdateResponse) Descriptor() ([]byte, []int) {
	return file_proto_server_pr_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type ValueRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metricsname string `protobuf:"bytes,1,opt,name=metricsname,proto3" json:"metricsname,omitempty"`
}

func (x *ValueRequest) Reset() {
	*x = ValueRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_server_pr_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValueRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValueRequest) ProtoMessage() {}

func (x *ValueRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_server_pr_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValueRequest.ProtoReflect.Descriptor instead.
func (*ValueRequest) Descriptor() ([]byte, []int) {
	return file_proto_server_pr_proto_rawDescGZIP(), []int{3}
}

func (x *ValueRequest) GetMetricsname() string {
	if x != nil {
		return x.Metricsname
	}
	return ""
}

type ValueResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metrics *Metrics `protobuf:"bytes,1,opt,name=metrics,proto3" json:"metrics,omitempty"`
}

func (x *ValueResponse) Reset() {
	*x = ValueResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_server_pr_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValueResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValueResponse) ProtoMessage() {}

func (x *ValueResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_server_pr_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValueResponse.ProtoReflect.Descriptor instead.
func (*ValueResponse) Descriptor() ([]byte, []int) {
	return file_proto_server_pr_proto_rawDescGZIP(), []int{4}
}

func (x *ValueResponse) GetMetrics() *Metrics {
	if x != nil {
		return x.Metrics
	}
	return nil
}

var File_proto_server_pr_proto protoreflect.FileDescriptor

var file_proto_server_pr_proto_rawDesc = []byte{
	0x0a, 0x15, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x70,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0xa5, 0x01,
	0x0a, 0x07, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x29, 0x0a, 0x05, 0x6d, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d, 0x54, 0x79, 0x70, 0x65, 0x52, 0x05, 0x6d,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x05, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x68, 0x61, 0x73, 0x68, 0x22, 0x1f, 0x0a, 0x05, 0x4d, 0x54, 0x79, 0x70, 0x65, 0x12, 0x09, 0x0a,
	0x05, 0x47, 0x41, 0x55, 0x47, 0x45, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x55, 0x4e,
	0x54, 0x45, 0x52, 0x10, 0x01, 0x22, 0x38, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x22,
	0x26, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x30, 0x0a, 0x0c, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x6d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x38, 0x0a, 0x0d, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x07, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6d, 0x61,
	0x69, 0x6e, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x07, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x32, 0x6d, 0x0a, 0x04, 0x47, 0x72, 0x70, 0x63, 0x12, 0x33, 0x0a, 0x06, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x13, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x6d, 0x61, 0x69,
	0x6e, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x30, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x12, 0x2e, 0x6d, 0x61, 0x69, 0x6e,
	0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e,
	0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x04, 0x5a, 0x02, 0x2e, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_server_pr_proto_rawDescOnce sync.Once
	file_proto_server_pr_proto_rawDescData = file_proto_server_pr_proto_rawDesc
)

func file_proto_server_pr_proto_rawDescGZIP() []byte {
	file_proto_server_pr_proto_rawDescOnce.Do(func() {
		file_proto_server_pr_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_server_pr_proto_rawDescData)
	})
	return file_proto_server_pr_proto_rawDescData
}

var file_proto_server_pr_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_server_pr_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_server_pr_proto_goTypes = []interface{}{
	(Metrics_MType)(0),     // 0: main.Metrics.MType
	(*Metrics)(nil),        // 1: main.Metrics
	(*UpdateRequest)(nil),  // 2: main.UpdateRequest
	(*UpdateResponse)(nil), // 3: main.UpdateResponse
	(*ValueRequest)(nil),   // 4: main.ValueRequest
	(*ValueResponse)(nil),  // 5: main.ValueResponse
}
var file_proto_server_pr_proto_depIdxs = []int32{
	0, // 0: main.Metrics.mtype:type_name -> main.Metrics.MType
	1, // 1: main.UpdateRequest.metrics:type_name -> main.Metrics
	1, // 2: main.ValueResponse.metrics:type_name -> main.Metrics
	2, // 3: main.Grpc.Update:input_type -> main.UpdateRequest
	4, // 4: main.Grpc.Value:input_type -> main.ValueRequest
	3, // 5: main.Grpc.Update:output_type -> main.UpdateResponse
	5, // 6: main.Grpc.Value:output_type -> main.ValueResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proto_server_pr_proto_init() }
func file_proto_server_pr_proto_init() {
	if File_proto_server_pr_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_server_pr_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metrics); i {
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
		file_proto_server_pr_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRequest); i {
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
		file_proto_server_pr_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateResponse); i {
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
		file_proto_server_pr_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValueRequest); i {
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
		file_proto_server_pr_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValueResponse); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_server_pr_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_server_pr_proto_goTypes,
		DependencyIndexes: file_proto_server_pr_proto_depIdxs,
		EnumInfos:         file_proto_server_pr_proto_enumTypes,
		MessageInfos:      file_proto_server_pr_proto_msgTypes,
	}.Build()
	File_proto_server_pr_proto = out.File
	file_proto_server_pr_proto_rawDesc = nil
	file_proto_server_pr_proto_goTypes = nil
	file_proto_server_pr_proto_depIdxs = nil
}