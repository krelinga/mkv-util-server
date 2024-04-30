package main

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "log"
    "net"
    "os"
    "os/exec"

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

func (s *MkvUtilsServer) RunMkvToolNixCommand(ctx context.Context, r *pb.RunMkvToolNixCommandRequest) (*pb.RunMkvToolNixCommandReply, error) {
    var command string;
    switch r.Command {
    case pb.RunMkvToolNixCommandRequest_COMMAND_MKVINFO:
        command = "mkvinfo"
    default:
        return nil, fmt.Errorf("Unsupported command: %v", r.Command)
    }
    cmd := exec.CommandContext(ctx, command, r.Args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    getExitCode := func() int32 {
        if err == nil {
            return 0
        }

        var exitErr *exec.ExitError
        if errors.Is(err, exitErr) {
            return int32(exitErr.ExitCode())
        }

        return -1
    }
    return &pb.RunMkvToolNixCommandReply{
        ExitCode: getExitCode(),
        Stdout: stdout.String(),
        Stderr: stderr.String(),
    }, err
}

func (s *MkvUtilsServer) Concat(ctx context.Context, r *pb.ConcatRequest) (*pb.ConcatReply, error) {
    return concat(ctx, r)
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
