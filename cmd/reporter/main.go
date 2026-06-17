package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/TheodoreQQ/polls-go/internal/repository"
	"github.com/TheodoreQQ/polls-go/pb"
)

type server struct {
	pb.UnimplementedReportServiceServer
	repo *repository.PollRepository
}

func (s *server) GenerateReport(ctx context.Context, req *pb.ExportRequest) (*pb.ExportResponse, error) {
	log.Printf("Generating raport for poll with ID: %d", req.PollId)

	md, _ := metadata.FromIncomingContext(ctx)
	currentUserID := md.Get("x-user-id")[0]

	poll, err := s.repo.GetResultsForBroadcast(int(req.PollId))
	if err != nil {
		return nil, status.Error(codes.NotFound, "Poll does not exist")
	}

	if fmt.Sprintf("%v", poll.OwnerID) != currentUserID {
		log.Printf("You have no permission to download that poll")
		return nil, status.Error(codes.PermissionDenied, "No permisson")
	}

	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)
	writer.Comma = ';'

	writer.Write([]string{"Pytanie", poll.Question})
	writer.Write([]string{"Opcja", "Liczba glosow"})

	for _, opt := range poll.Options {
		row := []string{
			opt.Text,
			fmt.Sprintf("%d", opt.VotesCount),
		}
		writer.Write(row)
	}

	writer.Write([]string{"SUMA", fmt.Sprintf("%d", poll.TotalVotes)})

	writer.Flush()

	return &pb.ExportResponse{
		Content:  b.Bytes(),
		Filename: fmt.Sprintf("ankieta_%d.csv", poll.ID),
	}, nil
}

func main() {
	// Connecting to database
	_ = godotenv.Load("../../.env")
	// if err != nil {
	// 	log.Fatal("Failed to load .env file")
	// }

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		log.Fatal("URL is empty")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Repo intialization
	pollRepo := repository.NewPollRepository(db)

	// Turning on gRCP
	port := os.Getenv("REPORTER_PORT")
	if port == "" {
		port = "50051"
	}

	addr := ":" + port

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterReportServiceServer(s, &server{repo: pollRepo})

	log.Println("Mikroserwis gRPC Reporter śmiga na porcie :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Błąd serwera: %v", err)
	}
}
