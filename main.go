package main

import (
    "log"
    "net"

    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/grpc"
)

type MkvUtilsServer struct {
    pb.UnimplementedMkvUtilsServer
}

func MainOrError() error {
    lis, err := net.Listen("tcp", ":25002")
    if err != nil {
        return err
    }
    grpcServer := grpc.NewServer()
    pb.RegisterMkvUtilsServer(grpcServer, &MkvUtilsServer{})
    grpcServer.Serve(lis)  // Runs as long as the server is alive.

    return nil
}

func main() {
    if err := MainOrError(); err != nil {
        log.Fatal(err)
    }
}
