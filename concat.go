package main

import (
    "context"
    "os/exec"

    "github.com/krelinga/mkv-util-server/pb"
)

func concat(ctx context.Context, r *pb.ConcatRequest) (*pb.ConcatReply, error) {
    args := []string{
        "-o", r.OutputPath,
    }
    for i, p := range r.InputPaths {
        if i > 0 {
            args = append(args, "+")
        }
        args = append(args, p)
    }
    cmd := exec.CommandContext(ctx, "mkvmerge", args...)
    if err := cmd.Run(); err != nil {
        return nil, err
    }
    return &pb.ConcatReply{}, nil
}
