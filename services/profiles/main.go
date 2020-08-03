package main

import (
	ps "./proto"

	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type ProfilesServiceServer struct{}

func (s *ProfilesServiceServer) SignUp(ctx context.Context,
	req *ps.SignUpRequest) (*ps.SignUpResponse, error) {

	response := new(ps.SignUpResponse)

	response.IsOK = true
	response.Description = "OK"

	return response, nil
}

func main() {
	server := grpc.NewServer()
	instance := new(ProfilesServiceServer)
	ps.RegisterProfilesServiceServer(server, instance)

	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal("Unable to create grpc listener", err)
	}

	if err = server.Serve(listener); err != nil {
		log.Fatal("Unable to start server", err)
	}
}
