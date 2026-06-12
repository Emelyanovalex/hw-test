package internalgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	api "github.com/Emelyanovalex/hw12_calendar/api"
	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error)
}

type Config struct {
	Host string
	Port int
}

type Server struct {
	server *grpc.Server
	logger Logger
	addr   string
}

func NewServer(logger Logger, app Application, cfg Config) *Server {
	interceptor := func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			logger.Error(fmt.Sprintf("grpc %s error: %v", info.FullMethod, err))
		} else {
			logger.Info(fmt.Sprintf("grpc %s ok", info.FullMethod))
		}
		return resp, err
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	api.RegisterCalendarServiceServer(grpcServer, &calendarService{app: app})

	return &Server{
		server: grpcServer,
		logger: logger,
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

func (s *Server) Start(_ context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.logger.Info(fmt.Sprintf("grpc server listening on %s", s.addr))
	return s.server.Serve(lis)
}

func (s *Server) Stop(_ context.Context) error {
	s.server.GracefulStop()
	s.logger.Info("grpc server stopped")
	return nil
}

type calendarService struct {
	api.UnimplementedCalendarServiceServer
	app Application
}

func (c *calendarService) CreateEvent(ctx context.Context, req *api.CreateEventRequest) (*api.CreateEventResponse, error) {
	event := protoToEvent(req.GetEvent())
	if err := c.app.CreateEvent(ctx, event); err != nil {
		return nil, toGRPCError(err)
	}
	return &api.CreateEventResponse{Event: eventToProto(event)}, nil
}

func (c *calendarService) UpdateEvent(ctx context.Context, req *api.UpdateEventRequest) (*api.UpdateEventResponse, error) {
	event := protoToEvent(req.GetEvent())
	if err := c.app.UpdateEvent(ctx, req.GetId(), event); err != nil {
		return nil, toGRPCError(err)
	}
	event.ID = req.GetId()
	return &api.UpdateEventResponse{Event: eventToProto(event)}, nil
}

func (c *calendarService) DeleteEvent(ctx context.Context, req *api.DeleteEventRequest) (*api.DeleteEventResponse, error) {
	if err := c.app.DeleteEvent(ctx, req.GetId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &api.DeleteEventResponse{}, nil
}

func (c *calendarService) ListEventsForDay(ctx context.Context, req *api.ListEventsRequest) (*api.ListEventsResponse, error) {
	events, err := c.app.ListEventsForDay(ctx, time.Unix(req.GetDate(), 0).UTC())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &api.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func (c *calendarService) ListEventsForWeek(ctx context.Context, req *api.ListEventsRequest) (*api.ListEventsResponse, error) {
	events, err := c.app.ListEventsForWeek(ctx, time.Unix(req.GetDate(), 0).UTC())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &api.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func (c *calendarService) ListEventsForMonth(ctx context.Context, req *api.ListEventsRequest) (*api.ListEventsResponse, error) {
	events, err := c.app.ListEventsForMonth(ctx, time.Unix(req.GetDate(), 0).UTC())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &api.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func protoToEvent(p *api.Event) storage.Event {
	if p == nil {
		return storage.Event{}
	}
	return storage.Event{
		ID:           p.GetId(),
		Title:        p.GetTitle(),
		StartTime:    time.Unix(p.GetStartTime(), 0).UTC(),
		Duration:     time.Duration(p.GetDuration()),
		Description:  p.GetDescription(),
		UserID:       p.GetUserId(),
		NotifyBefore: time.Duration(p.GetNotifyBefore()),
	}
}

func eventToProto(e storage.Event) *api.Event {
	return &api.Event{
		Id:           e.ID,
		Title:        e.Title,
		StartTime:    e.StartTime.Unix(),
		Duration:     int64(e.Duration),
		Description:  e.Description,
		UserId:       e.UserID,
		NotifyBefore: int64(e.NotifyBefore),
	}
}

func eventsToProto(events []storage.Event) []*api.Event {
	result := make([]*api.Event, 0, len(events))
	for _, e := range events {
		result = append(result, eventToProto(e))
	}
	return result
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, storage.ErrDateBusy):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
