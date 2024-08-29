package handlers

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	Service application.CandlestickService
	pb.UnimplementedCandlestickServiceServer
}

func (s *GRPCServer) BroadcastCandlestick(ctx context.Context, data *pb.CandlestickData) (*pb.CandlestickResponse, error) {
	log.Printf("Received candlestick: %+v", data)
	return &pb.CandlestickResponse{Success: true}, nil
}

func StartGRPCServer(service application.CandlestickService) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCandlestickServiceServer(s, &GRPCServer{Service: service})
	log.Println("gRPC server listening at", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
