// Code generated by protoc-gen-go.
// source: api.proto
// DO NOT EDIT!

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	api.proto
	types.proto

It has these top-level messages:
	EventsRequest
	EventsResponse
	Filters
	Node
	Service
	State
	Event
*/
package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type EventsRequest struct {
	Filters *Filters `protobuf:"bytes,1,opt,name=Filters,json=filters" json:"Filters,omitempty"`
}

func (m *EventsRequest) Reset()                    { *m = EventsRequest{} }
func (m *EventsRequest) String() string            { return proto.CompactTextString(m) }
func (*EventsRequest) ProtoMessage()               {}
func (*EventsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *EventsRequest) GetFilters() *Filters {
	if m != nil {
		return m.Filters
	}
	return nil
}

type EventsResponse struct {
	Event *Event `protobuf:"bytes,1,opt,name=Event,json=event" json:"Event,omitempty"`
}

func (m *EventsResponse) Reset()                    { *m = EventsResponse{} }
func (m *EventsResponse) String() string            { return proto.CompactTextString(m) }
func (*EventsResponse) ProtoMessage()               {}
func (*EventsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *EventsResponse) GetEvent() *Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func init() {
	proto.RegisterType((*EventsRequest)(nil), "api.EventsRequest")
	proto.RegisterType((*EventsResponse)(nil), "api.EventsResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Hypochronos service

type HypochronosClient interface {
	Events(ctx context.Context, in *EventsRequest, opts ...grpc.CallOption) (Hypochronos_EventsClient, error)
}

type hypochronosClient struct {
	cc *grpc.ClientConn
}

func NewHypochronosClient(cc *grpc.ClientConn) HypochronosClient {
	return &hypochronosClient{cc}
}

func (c *hypochronosClient) Events(ctx context.Context, in *EventsRequest, opts ...grpc.CallOption) (Hypochronos_EventsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Hypochronos_serviceDesc.Streams[0], c.cc, "/api.Hypochronos/Events", opts...)
	if err != nil {
		return nil, err
	}
	x := &hypochronosEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Hypochronos_EventsClient interface {
	Recv() (*EventsResponse, error)
	grpc.ClientStream
}

type hypochronosEventsClient struct {
	grpc.ClientStream
}

func (x *hypochronosEventsClient) Recv() (*EventsResponse, error) {
	m := new(EventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Hypochronos service

type HypochronosServer interface {
	Events(*EventsRequest, Hypochronos_EventsServer) error
}

func RegisterHypochronosServer(s *grpc.Server, srv HypochronosServer) {
	s.RegisterService(&_Hypochronos_serviceDesc, srv)
}

func _Hypochronos_Events_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(EventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HypochronosServer).Events(m, &hypochronosEventsServer{stream})
}

type Hypochronos_EventsServer interface {
	Send(*EventsResponse) error
	grpc.ServerStream
}

type hypochronosEventsServer struct {
	grpc.ServerStream
}

func (x *hypochronosEventsServer) Send(m *EventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _Hypochronos_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.Hypochronos",
	HandlerType: (*HypochronosServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Events",
			Handler:       _Hypochronos_Events_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "api.proto",
}

func init() { proto.RegisterFile("api.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 165 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4c, 0x2c, 0xc8, 0xd4,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4e, 0x2c, 0xc8, 0x94, 0xe2, 0x2e, 0xa9, 0x2c, 0x48,
	0x2d, 0x86, 0x88, 0x28, 0x99, 0x73, 0xf1, 0xba, 0x96, 0xa5, 0xe6, 0x95, 0x14, 0x07, 0xa5, 0x16,
	0x96, 0xa6, 0x16, 0x97, 0x08, 0xa9, 0x71, 0xb1, 0xbb, 0x65, 0xe6, 0x94, 0xa4, 0x16, 0x15, 0x4b,
	0x30, 0x2a, 0x30, 0x6a, 0x70, 0x1b, 0xf1, 0xe8, 0x81, 0xf4, 0x43, 0xc5, 0x82, 0xd8, 0xd3, 0x20,
	0x0c, 0x25, 0x23, 0x2e, 0x3e, 0x98, 0xc6, 0xe2, 0x82, 0xfc, 0xbc, 0xe2, 0x54, 0x21, 0x05, 0x2e,
	0x56, 0xb0, 0x08, 0x54, 0x1f, 0x17, 0x58, 0x1f, 0x58, 0x24, 0x88, 0x35, 0x15, 0x44, 0x19, 0x39,
	0x71, 0x71, 0x7b, 0x54, 0x16, 0xe4, 0x27, 0x67, 0x14, 0xe5, 0xe7, 0xe5, 0x17, 0x0b, 0x19, 0x73,
	0xb1, 0x41, 0x8c, 0x10, 0x12, 0x42, 0xa8, 0x85, 0x39, 0x44, 0x4a, 0x18, 0x45, 0x0c, 0x62, 0x87,
	0x01, 0x63, 0x12, 0x1b, 0xd8, 0xdd, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xaa, 0x13, 0x3e,
	0x46, 0xd6, 0x00, 0x00, 0x00,
}
