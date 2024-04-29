package main

import (
    "context"
    "log"
    "net"
    "os"

    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/grpc"
)

type MkvUtilsServer struct {
    pb.UnimplementedMkvUtilsServer
}

func (s *MkvUtilsServer) GetFileSize(_ context.Context, r *pb.GetFileSizeRequest) (*pb.GetFileSizeReply, error) {
    stat, err := os.Stat(r.Path)
    if err != nil {
        return nil, err
    }
    return &pb.GetFileSizeReply{
        Size: stat.Size(),
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
