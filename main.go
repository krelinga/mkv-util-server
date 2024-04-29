package main

import (
    "context"
    "log"
    "net"

    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/grpc"
)

type MkvUtilsServer struct {
    pb.UnimplementedMkvUtilsServer
}

func (s *MkvUtilsServer) GetFileSize(_ context.Context, r *pb.GetFileSizeRequest) (*pb.GetFileSizeReply, error) {
    return &pb.GetFileSizeReply{
        Size: -1,
    }, nil
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
