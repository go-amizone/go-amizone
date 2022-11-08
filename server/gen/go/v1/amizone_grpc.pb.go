// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: v1/amizone.proto

package api_v1

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

// AmizoneServiceClient is the client API for AmizoneService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AmizoneServiceClient interface {
	GetAttendance(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*AttendanceRecords, error)
	GetClassSchedule(ctx context.Context, in *ClassScheduleRequest, opts ...grpc.CallOption) (*ScheduledClasses, error)
	// GetExamSchedule returns exam schedule. Amizone only allows access to schedules for the ongoing semester
	// and only close to the exam dates, so we don't take any parameters.
	GetExamSchedule(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*ExaminationSchedule, error)
	// GetSemesters returns a list of semesters that include past semesters and the current semester.
	// These semesters can be used in other RPCs that consume them, for example GetCourses.
	GetSemesters(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*SemesterList, error)
	// GetCourses returns a list of courses for the given semester.
	GetCourses(ctx context.Context, in *SemesterRef, opts ...grpc.CallOption) (*Courses, error)
	// GetCurrentCourses returns a list of courses for the "current" semester.
	GetCurrentCourses(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*Courses, error)
	// GetUserProfile returns the user's profile.
	GetUserProfile(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*Profile, error)
}

type amizoneServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAmizoneServiceClient(cc grpc.ClientConnInterface) AmizoneServiceClient {
	return &amizoneServiceClient{cc}
}

func (c *amizoneServiceClient) GetAttendance(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*AttendanceRecords, error) {
	out := new(AttendanceRecords)
	err := c.cc.Invoke(ctx, "/go_amizone.server.proto.v1.AmizoneService/GetAttendance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *amizoneServiceClient) GetClassSchedule(ctx context.Context, in *ClassScheduleRequest, opts ...grpc.CallOption) (*ScheduledClasses, error) {
	out := new(ScheduledClasses)
	err := c.cc.Invoke(ctx, "/go_amizone.server.proto.v1.AmizoneService/GetClassSchedule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *amizoneServiceClient) GetExamSchedule(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*ExaminationSchedule, error) {
	out := new(ExaminationSchedule)
	err := c.cc.Invoke(ctx, "/go_amizone.server.proto.v1.AmizoneService/GetExamSchedule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *amizoneServiceClient) GetSemesters(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*SemesterList, error) {
	out := new(SemesterList)
	err := c.cc.Invoke(ctx, "/go_amizone.server.proto.v1.AmizoneService/GetSemesters", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *amizoneServiceClient) GetCourses(ctx context.Context, in *SemesterRef, opts ...grpc.CallOption) (*Courses, error) {
	out := new(Courses)
	err := c.cc.Invoke(ctx, "/go_amizone.server.proto.v1.AmizoneService/GetCourses", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *amizoneServiceClient) GetCurrentCourses(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*Courses, error) {
	out := new(Courses)
	err := c.cc.Invoke(ctx, "/go_amizone.server.proto.v1.AmizoneService/GetCurrentCourses", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *amizoneServiceClient) GetUserProfile(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*Profile, error) {
	out := new(Profile)
	err := c.cc.Invoke(ctx, "/go_amizone.server.proto.v1.AmizoneService/GetUserProfile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AmizoneServiceServer is the server API for AmizoneService service.
// All implementations must embed UnimplementedAmizoneServiceServer
// for forward compatibility
type AmizoneServiceServer interface {
	GetAttendance(context.Context, *EmptyMessage) (*AttendanceRecords, error)
	GetClassSchedule(context.Context, *ClassScheduleRequest) (*ScheduledClasses, error)
	// GetExamSchedule returns exam schedule. Amizone only allows access to schedules for the ongoing semester
	// and only close to the exam dates, so we don't take any parameters.
	GetExamSchedule(context.Context, *EmptyMessage) (*ExaminationSchedule, error)
	// GetSemesters returns a list of semesters that include past semesters and the current semester.
	// These semesters can be used in other RPCs that consume them, for example GetCourses.
	GetSemesters(context.Context, *EmptyMessage) (*SemesterList, error)
	// GetCourses returns a list of courses for the given semester.
	GetCourses(context.Context, *SemesterRef) (*Courses, error)
	// GetCurrentCourses returns a list of courses for the "current" semester.
	GetCurrentCourses(context.Context, *EmptyMessage) (*Courses, error)
	// GetUserProfile returns the user's profile.
	GetUserProfile(context.Context, *EmptyMessage) (*Profile, error)
	mustEmbedUnimplementedAmizoneServiceServer()
}

// UnimplementedAmizoneServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAmizoneServiceServer struct {
}

func (UnimplementedAmizoneServiceServer) GetAttendance(context.Context, *EmptyMessage) (*AttendanceRecords, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAttendance not implemented")
}
func (UnimplementedAmizoneServiceServer) GetClassSchedule(context.Context, *ClassScheduleRequest) (*ScheduledClasses, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClassSchedule not implemented")
}
func (UnimplementedAmizoneServiceServer) GetExamSchedule(context.Context, *EmptyMessage) (*ExaminationSchedule, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetExamSchedule not implemented")
}
func (UnimplementedAmizoneServiceServer) GetSemesters(context.Context, *EmptyMessage) (*SemesterList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSemesters not implemented")
}
func (UnimplementedAmizoneServiceServer) GetCourses(context.Context, *SemesterRef) (*Courses, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCourses not implemented")
}
func (UnimplementedAmizoneServiceServer) GetCurrentCourses(context.Context, *EmptyMessage) (*Courses, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCurrentCourses not implemented")
}
func (UnimplementedAmizoneServiceServer) GetUserProfile(context.Context, *EmptyMessage) (*Profile, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserProfile not implemented")
}
func (UnimplementedAmizoneServiceServer) mustEmbedUnimplementedAmizoneServiceServer() {}

// UnsafeAmizoneServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AmizoneServiceServer will
// result in compilation errors.
type UnsafeAmizoneServiceServer interface {
	mustEmbedUnimplementedAmizoneServiceServer()
}

func RegisterAmizoneServiceServer(s grpc.ServiceRegistrar, srv AmizoneServiceServer) {
	s.RegisterService(&AmizoneService_ServiceDesc, srv)
}

func _AmizoneService_GetAttendance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AmizoneServiceServer).GetAttendance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_amizone.server.proto.v1.AmizoneService/GetAttendance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AmizoneServiceServer).GetAttendance(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _AmizoneService_GetClassSchedule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClassScheduleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AmizoneServiceServer).GetClassSchedule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_amizone.server.proto.v1.AmizoneService/GetClassSchedule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AmizoneServiceServer).GetClassSchedule(ctx, req.(*ClassScheduleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AmizoneService_GetExamSchedule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AmizoneServiceServer).GetExamSchedule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_amizone.server.proto.v1.AmizoneService/GetExamSchedule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AmizoneServiceServer).GetExamSchedule(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _AmizoneService_GetSemesters_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AmizoneServiceServer).GetSemesters(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_amizone.server.proto.v1.AmizoneService/GetSemesters",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AmizoneServiceServer).GetSemesters(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _AmizoneService_GetCourses_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SemesterRef)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AmizoneServiceServer).GetCourses(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_amizone.server.proto.v1.AmizoneService/GetCourses",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AmizoneServiceServer).GetCourses(ctx, req.(*SemesterRef))
	}
	return interceptor(ctx, in, info, handler)
}

func _AmizoneService_GetCurrentCourses_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AmizoneServiceServer).GetCurrentCourses(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_amizone.server.proto.v1.AmizoneService/GetCurrentCourses",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AmizoneServiceServer).GetCurrentCourses(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _AmizoneService_GetUserProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AmizoneServiceServer).GetUserProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_amizone.server.proto.v1.AmizoneService/GetUserProfile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AmizoneServiceServer).GetUserProfile(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// AmizoneService_ServiceDesc is the grpc.ServiceDesc for AmizoneService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AmizoneService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "go_amizone.server.proto.v1.AmizoneService",
	HandlerType: (*AmizoneServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAttendance",
			Handler:    _AmizoneService_GetAttendance_Handler,
		},
		{
			MethodName: "GetClassSchedule",
			Handler:    _AmizoneService_GetClassSchedule_Handler,
		},
		{
			MethodName: "GetExamSchedule",
			Handler:    _AmizoneService_GetExamSchedule_Handler,
		},
		{
			MethodName: "GetSemesters",
			Handler:    _AmizoneService_GetSemesters_Handler,
		},
		{
			MethodName: "GetCourses",
			Handler:    _AmizoneService_GetCourses_Handler,
		},
		{
			MethodName: "GetCurrentCourses",
			Handler:    _AmizoneService_GetCurrentCourses_Handler,
		},
		{
			MethodName: "GetUserProfile",
			Handler:    _AmizoneService_GetUserProfile_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/amizone.proto",
}
