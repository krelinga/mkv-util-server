package main

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "net/http"
    "os"
    "os/exec"

    "connectrpc.com/connect"
    "golang.org/x/net/http2"
    "golang.org/x/net/http2/h2c"
    
    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
    pbconnect "buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video/mkv_util_server/v1/mkv_util_serverv1connect"
)

type MkvUtilServer struct {
}

func (s *MkvUtilServer) GetFileSize(_ context.Context, r *connect.Request[pb.GetFileSizeRequest]) (*connect.Response[pb.GetFileSizeResponse], error) {
    stat, err := os.Stat(r.Msg.Path)
    if err != nil {
        return nil, err
    }
    return connect.NewResponse(&pb.GetFileSizeResponse{
        Size: stat.Size(),
    }), nil
}

func (s *MkvUtilServer) RunMkvToolNixCommand(ctx context.Context, r *connect.Request[pb.RunMkvToolNixCommandRequest]) (*connect.Response[pb.RunMkvToolNixCommandResponse], error) {
    var command string;
    switch r.Msg.Command {
    case pb.RunMkvToolNixCommandRequest_COMMAND_MKVINFO:
        command = "mkvinfo"
    default:
        return nil, fmt.Errorf("Unsupported command: %v", r.Msg.Command)
    }
    cmd := exec.CommandContext(ctx, command, r.Msg.Args...)
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
    return connect.NewResponse(&pb.RunMkvToolNixCommandResponse{
        ExitCode: getExitCode(),
        Stdout: stdout.String(),
        Stderr: stderr.String(),
    }), err
}

func (s *MkvUtilServer) Concat(ctx context.Context, r *connect.Request[pb.ConcatRequest]) (*connect.Response[pb.ConcatResponse], error) {
    resp, err := concat(ctx, r.Msg)
    if err != nil {
        return nil, err
    }
    return connect.NewResponse(resp), nil
}

func (s *MkvUtilServer) GetChapters(ctx context.Context, r *connect.Request[pb.GetChaptersRequest]) (*connect.Response[pb.GetChaptersResponse], error) {
    resp, err := getChapters(ctx, r.Msg)
    if err != nil {
        return nil, err
    }
    return connect.NewResponse(resp), nil
}

func (s *MkvUtilServer) GetInfo(ctx context.Context, r *connect.Request[pb.GetInfoRequest]) (*connect.Response[pb.GetInfoResponse], error) {
    resp, err := getInfo(ctx, r.Msg)
    if err != nil {
        return nil, err
    }
    return connect.NewResponse(resp), nil
}

func (s *MkvUtilServer) Split(ctx context.Context, r *connect.Request[pb.SplitRequest]) (*connect.Response[pb.SplitResponse], error) {
    resp, err := split(ctx, r.Msg)
    if err != nil {
        return nil, err
    }
    return connect.NewResponse(resp), nil
}

func main() {
    mux := http.NewServeMux()
    path, handler := pbconnect.NewMkvUtilServiceHandler(&MkvUtilServer{})
    mux.Handle(path, handler)
    // Runs as long as the server is alive.
    http.ListenAndServe("0.0.0.0:25002", h2c.NewHandler(mux, &http2.Server{}))
}
