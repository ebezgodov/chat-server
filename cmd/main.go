package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/brianvoe/gofakeit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/ebezgodov/chat-server/pkg/chat_v1"
)

const grpcPort = 50052

type server struct {
	desc.UnimplementedChatV1Server
}

// Create ...
func (s *server) Create(_ context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	//Выводим в лог все элементы из списка repeated string usernames
	for _, username := range req.GetUsernames() {
		log.Printf("username: %s", username)
	}

	return &desc.CreateResponse{
		ChatId: gofakeit.Int64(),
	}, nil
}

// Delete ...
func (s *server) Delete(_ context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("chat id: %d", req.GetChatId())

	return new(emptypb.Empty), nil
}

// SendMessage ...
func (s *server) SendMessage(_ context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	log.Printf("from: %s, text: %s, timestamp: %s", req.GetFrom(), req.GetText(), req.GetTimestamp())

	return new(emptypb.Empty), nil
}

// Main
func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
