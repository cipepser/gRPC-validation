package main

import (
	"context"
	"errors"
	"log"
	"net"

	pb "github.com/cipepser/gRPC-validation/user"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
)

type server struct {
	users map[*pb.User]struct{}
	names map[string]struct{}
}

const (
	port = ":50051"
)

var (
	empty = new(pb.Empty)
)

func (s *server) GetUser(ctx context.Context, in *pb.Name) (*pb.User, error) {
	for u := range s.users {
		if u.Name == in.Name {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *server) GetUsers(ctx context.Context, in *pb.Empty) (*pb.Users, error) {
	out := new(pb.Users)
	for u := range s.users {
		out.Users = append(out.Users, u)
	}
	return out, nil
}

func (s *server) AddUser(ctx context.Context, in *pb.User) (*pb.Empty, error) {
	if _, ok := s.names[in.Name]; ok {
		return empty, errors.New("user already exists")
	}
	s.users[in] = struct{}{}
	s.names[in.Name] = struct{}{}
	return empty, nil
}

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_validator.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_validator.UnaryServerInterceptor(),
		)),
	)
	pb.RegisterUserServiceServer(s,
		&server{
			users: map[*pb.User]struct{}{},
			names: map[string]struct{}{},
		},
	)
	s.Serve(l)
}
