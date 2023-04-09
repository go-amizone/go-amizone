package server

import (
	"context"
	"errors"
	"net"

	"github.com/ditsuke/go-amizone/amizone"
	v1 "github.com/ditsuke/go-amizone/server/gen/go/v1"
	"github.com/ditsuke/go-amizone/server/transformers/fromproto"
	"github.com/ditsuke/go-amizone/server/transformers/toproto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// serviceServer is an implementation of v1.AmizoneServiceServer. Plugged into proto-generated code, this
// implementation makes the Amizone API available over gRPC.
type serviceServer struct {
	v1.UnimplementedAmizoneServiceServer
}

func NewAmizoneServiceServer() v1.AmizoneServiceServer {
	return &serviceServer{}
}

func (a *serviceServer) GetAttendance(ctx context.Context, _ *v1.EmptyMessage) (*v1.AttendanceRecords, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to authenticate")
	}

	attendance, err := amizoneClient.GetAttendance()
	if err != nil {
		return nil, errors.New("failed to retrieve attendance")
	}

	return toproto.AttendanceRecords(attendance), nil
}

func (a serviceServer) GetClassSchedule(ctx context.Context, in *v1.ClassScheduleRequest) (*v1.ScheduledClasses, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to authenticate")
	}

	pDate := in.GetDate()
	if pDate == nil {
		return nil, status.Errorf(codes.InvalidArgument, "date is required")
	}
	nDate := fromproto.Date(pDate)
	schedule, err := amizoneClient.GetClassSchedule(nDate.Date())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve class schedule: %v", err)
	}

	return toproto.ScheduledClasses(schedule), nil
}

func (serviceServer) GetExamSchedule(ctx context.Context, _ *v1.EmptyMessage) (*v1.ExaminationSchedule, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to authenticate")
	}

	schedule, err := amizoneClient.GetExamSchedule()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve exam schedule: %v", err)
	}

	return toproto.ExamSchedule(*schedule), nil
}

func (serviceServer) GetSemesters(ctx context.Context, _ *v1.EmptyMessage) (*v1.SemesterList, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to authenticate")
	}

	semesters, err := amizoneClient.GetSemesters()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve semesters: %v", err)
	}

	return toproto.SemesterList(semesters), nil
}

func (serviceServer) GetCourses(ctx context.Context, in *v1.SemesterRef) (*v1.Courses, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to authenticate")
	}

	if in.GetSemesterRef() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "semester ref is required")
	}

	courses, err := amizoneClient.GetCourses(in.GetSemesterRef())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve courses: %v", err)
	}

	return toproto.Courses(courses), nil
}

func (serviceServer) GetCurrentCourses(ctx context.Context, _ *v1.EmptyMessage) (*v1.Courses, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to authenticate")
	}

	courses, err := amizoneClient.GetCurrentCourses()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve courses: %v", err)
	}

	return toproto.Courses(courses), nil
}

func (serviceServer) GetUserProfile(ctx context.Context, _ *v1.EmptyMessage) (*v1.Profile, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to authenticate")
	}

	profile, err := amizoneClient.GetProfile()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user-profile: %v", err)
	}
	return toproto.Profile(*profile), nil
}

func (serviceServer) GetWifiMacInfo(ctx context.Context, _ *v1.EmptyMessage) (*v1.WifiMacInfo, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to authenticate")
	}

	macInfo, err := amizoneClient.GetWifiMacInfo()
	if err != nil {
		// TODO: ! reevalute these error codes, I get the feeling they shouldn't just be codes.Internal
		return nil, status.Errorf(codes.Internal, "failed to retrieve mac info")
	}
	return toproto.WifiInfo(*macInfo), nil
}

func (serviceServer) RegisterWifiMac(ctx context.Context, req *v1.RegisterWifiMacRequest) (*v1.EmptyMessage, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to authenticate")
	}
	addr, err := net.ParseMAC(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad mac address")
	}

	err = amizoneClient.RegisterWifiMac(addr, req.OverrideLimit)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to register: %s", err.Error())
	}

	return &v1.EmptyMessage{}, nil
}

func (serviceServer) DeregisterWifiMac(ctx context.Context, req *v1.DeregisterWifiMacRequest) (*v1.EmptyMessage, error) {
	amizoneClient, ok := ctx.Value(ContextAmizoneClientKey).(*amizone.Client)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to authenticate")
	}

	addr, err := net.ParseMAC(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad mac address")
	}
	err = amizoneClient.RemoveWifiMac(addr)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed removal: %s", err.Error())
	}

	return &v1.EmptyMessage{}, nil
}
