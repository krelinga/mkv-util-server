package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"

    "github.com/krelinga/mkv-util-server/pb"
)

func getChaptersSimple(ctx context.Context, path string) (*pb.SimpleChapters, error) {
    tmpDir, err := os.MkdirTemp("", "")
    if err != nil {
        return nil, fmt.Errorf("Could not create temporary directory: %e", err)
    }
    defer func() {
        if err := os.RemoveAll(tmpDir); err != nil {
            log.Printf("could not remove temp dir %s: %e", tmpDir, err)
        }
    }()
    return nil, errors.New("Not implemented")
}

func getChapters(ctx context.Context, r *pb.GetChaptersRequest) (*pb.GetChaptersReply, error) {
    resp := &pb.GetChaptersReply {}
    switch r.Format {
    case pb.ChaptersFormat_CF_SIMPLE:
        simple, err := getChaptersSimple(ctx, r.InPath)
        if err != nil {
            return nil, err
        }
        resp.Chapters = &pb.Chapters{
            Format: pb.ChaptersFormat_CF_SIMPLE,
            Simple: simple,
        }
    default:
        return nil, fmt.Errorf("Unsupported format: %v", r.Format)
    }

    return resp, nil
}
