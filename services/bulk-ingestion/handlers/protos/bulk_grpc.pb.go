// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.8
// source: bulk.proto

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// BulkClient is the client API for Bulk service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BulkClient interface {
	Projects(ctx context.Context, in *ProjectsRequest, opts ...grpc.CallOption) (*ProjectsResponse, error)
	Namespaces(ctx context.Context, in *NamespacesRequest, opts ...grpc.CallOption) (*NamespacesResponse, error)
	Articles(ctx context.Context, in *ArticlesRequest, opts ...grpc.CallOption) (*ArticlesResponse, error)
}

type bulkClient struct {
	cc grpc.ClientConnInterface
}

func NewBulkClient(cc grpc.ClientConnInterface) BulkClient {
	return &bulkClient{cc}
}

func (c *bulkClient) Projects(ctx context.Context, in *ProjectsRequest, opts ...grpc.CallOption) (*ProjectsResponse, error) {
	out := new(ProjectsResponse)
	err := c.cc.Invoke(ctx, "/bulk.Bulk/Projects", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bulkClient) Namespaces(ctx context.Context, in *NamespacesRequest, opts ...grpc.CallOption) (*NamespacesResponse, error) {
	out := new(NamespacesResponse)
	err := c.cc.Invoke(ctx, "/bulk.Bulk/Namespaces", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bulkClient) Articles(ctx context.Context, in *ArticlesRequest, opts ...grpc.CallOption) (*ArticlesResponse, error) {
	out := new(ArticlesResponse)
	err := c.cc.Invoke(ctx, "/bulk.Bulk/Articles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BulkServer is the server API for Bulk service.
// All implementations must embed UnimplementedBulkServer
// for forward compatibility
type BulkServer interface {
	Projects(context.Context, *ProjectsRequest) (*ProjectsResponse, error)
	Namespaces(context.Context, *NamespacesRequest) (*NamespacesResponse, error)
	Articles(context.Context, *ArticlesRequest) (*ArticlesResponse, error)
	mustEmbedUnimplementedBulkServer()
}

// UnimplementedBulkServer must be embedded to have forward compatible implementations.
type UnimplementedBulkServer struct {
}

func (UnimplementedBulkServer) Projects(context.Context, *ProjectsRequest) (*ProjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Projects not implemented")
}
func (UnimplementedBulkServer) Namespaces(context.Context, *NamespacesRequest) (*NamespacesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Namespaces not implemented")
}
func (UnimplementedBulkServer) Articles(context.Context, *ArticlesRequest) (*ArticlesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Articles not implemented")
}
func (UnimplementedBulkServer) mustEmbedUnimplementedBulkServer() {}

// UnsafeBulkServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BulkServer will
// result in compilation errors.
type UnsafeBulkServer interface {
	mustEmbedUnimplementedBulkServer()
}

func RegisterBulkServer(s grpc.ServiceRegistrar, srv BulkServer) {
	s.RegisterService(&Bulk_ServiceDesc, srv)
}

func _Bulk_Projects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BulkServer).Projects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bulk.Bulk/Projects",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BulkServer).Projects(ctx, req.(*ProjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bulk_Namespaces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NamespacesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BulkServer).Namespaces(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bulk.Bulk/Namespaces",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BulkServer).Namespaces(ctx, req.(*NamespacesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bulk_Articles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ArticlesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BulkServer).Articles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bulk.Bulk/Articles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BulkServer).Articles(ctx, req.(*ArticlesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Bulk_ServiceDesc is the grpc.ServiceDesc for Bulk service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Bulk_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "bulk.Bulk",
	HandlerType: (*BulkServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Projects",
			Handler:    _Bulk_Projects_Handler,
		},
		{
			MethodName: "Namespaces",
			Handler:    _Bulk_Namespaces_Handler,
		},
		{
			MethodName: "Articles",
			Handler:    _Bulk_Articles_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bulk.proto",
}
