package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	sq "github.com/Masterminds/squirrel"
	desc "github.com/ebezgodov/chat-server/pkg/chat_v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

const grpcPort = 50052

type server struct {
	desc.UnimplementedChatV1Server
	pool *pgxpool.Pool
}

// Create ...
func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	if req.GetUsernames() == nil {
		return nil, fmt.Errorf("usernames is required")
	}

	builderInsert := sq.Insert("chat").
		PlaceholderFormat(sq.Dollar).
		Columns("usernames").
		Values(req.GetUsernames()).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	var chatID int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&chatID)
	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}

	log.Printf("inserted chat with id: %d", chatID)

	return &desc.CreateResponse{
		ChatId: chatID,
	}, nil
}

// Delete ...
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	builderDeleteOne := sq.Delete("chat").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.GetChatId()})

	query, args, err := builderDeleteOne.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Fatalf("failed to delete user: %v", err)
	}

	log.Printf("deleted %d rows", res.RowsAffected())

	return new(emptypb.Empty), nil
}

// SendMessage ...
func (s *server) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	if req.GetFrom() == "" || req.GetText() == "" || req.GetTimestamp() == nil {
		return nil, fmt.Errorf("from, text and timestamp are required")
	}

	builderInsert := sq.Insert("msg").
		PlaceholderFormat(sq.Dollar).
		Columns("from_user", "msg_text", "created_at").
		Values(req.GetFrom(), req.GetText(), req.GetTimestamp().AsTime()).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	var msgID int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&msgID)
	if err != nil {
		log.Fatalf("failed to insert msg: %v", err)
	}

	log.Printf("inserted message with id: %d", msgID)

	return new(emptypb.Empty), nil
}

// Main
func main() {
	ctx := context.Background()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаем пул соединений с базой данных
	pool, err := pgxpool.Connect(ctx, "host=pg-chat-local port=5432 dbname=chat user=chat-user password=chat-password sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatV1Server(s, &server{pool: pool})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
