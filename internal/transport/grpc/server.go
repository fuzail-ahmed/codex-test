package grpcserver

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/fuzail-ahmed/codex-test/internal/repository"
	"github.com/fuzail-ahmed/codex-test/internal/service"
	todov1 "github.com/fuzail-ahmed/codex-test/shared/gen/todo/v1"
)

type Server struct {
	todov1.UnimplementedTodoServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) CreateTodo(ctx context.Context, req *todov1.CreateTodoRequest) (*todov1.CreateTodoResponse, error) {
	todo, err := s.svc.Create(ctx, service.CreateTodoInput{Title: req.GetTitle(), Description: req.GetDescription()})
	if err != nil {
		return nil, err
	}
	return &todov1.CreateTodoResponse{Todo: mapTodo(todo)}, nil
}

func (s *Server) BulkCreateTodos(ctx context.Context, req *todov1.BulkCreateTodosRequest) (*todov1.BulkCreateTodosResponse, error) {
	inputs := make([]service.CreateTodoInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		inputs = append(inputs, service.CreateTodoInput{Title: item.GetTitle(), Description: item.GetDescription()})
	}
	todos, err := s.svc.BulkCreate(ctx, inputs)
	if err != nil {
		return nil, err
	}
	return &todov1.BulkCreateTodosResponse{Todos: mapTodos(todos)}, nil
}

func (s *Server) GetTodo(ctx context.Context, req *todov1.GetTodoRequest) (*todov1.GetTodoResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}
	todo, err := s.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &todov1.GetTodoResponse{Todo: mapTodo(todo)}, nil
}

func (s *Server) ListTodos(ctx context.Context, req *todov1.ListTodosRequest) (*todov1.ListTodosResponse, error) {
	todos, err := s.svc.List(ctx, repository.ListFilter{Limit: int(req.GetLimit()), Offset: int(req.GetOffset())})
	if err != nil {
		return nil, err
	}
	return &todov1.ListTodosResponse{Todos: mapTodos(todos)}, nil
}

func (s *Server) UpdateTodo(ctx context.Context, req *todov1.UpdateTodoRequest) (*todov1.UpdateTodoResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}
	input := service.UpdateTodoInput{}
	if req.Title != "" {
		input.Title = &req.Title
	}
	if req.Description != "" {
		input.Description = &req.Description
	}
	if req.Status != todov1.Status_STATUS_UNSPECIFIED {
		status := mapStatus(req.Status)
		input.Status = &status
	}
	updated, err := s.svc.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	return &todov1.UpdateTodoResponse{Todo: mapTodo(updated)}, nil
}

func (s *Server) DeleteTodo(ctx context.Context, req *todov1.DeleteTodoRequest) (*todov1.DeleteTodoResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}
	deleted, err := s.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	return &todov1.DeleteTodoResponse{Deleted: deleted}, nil
}

func ListenAndServe(addr string, svc *service.Service, opts ...grpc.ServerOption) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	server := grpc.NewServer(opts...)
	todov1.RegisterTodoServiceServer(server, New(svc))
	return server, lis, nil
}
