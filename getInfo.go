package main

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"

    "github.com/krelinga/mkv-util-server/idjson"
    "google.golang.org/protobuf/types/known/durationpb"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
)

// Returns nil if no video tracks found.
func findFirstVideoTrack(j *idjson.MkvMerge) *idjson.Track {
    for _, t := range j.Tracks {
        if t.Type == "video" {
            return t
        }
    }
    return nil
}

func getInfo(ctx context.Context, r *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
    cmd := exec.CommandContext(ctx, "mkvmerge", "-J", r.InPath)
    b := bytes.Buffer{}
    cmd.Stdout = &b
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("Could not run mkvmerge: %w", err)
    }
    j, err := idjson.Parse(&b)
    if err != nil {
        return nil, fmt.Errorf("Could not parse mkvmerge output: %w", err)
    }
    d, err := j.Container.Properties.ParseDuration()
    if err != nil {
        return nil, fmt.Errorf("Could not convert mkvmerge output to a time.Duration: %w", err)
    }
    resp := &pb.GetInfoResponse{
        Info: &pb.Info{
            Duration: durationpb.New(d),
        },
    }
    return resp, nil
}
