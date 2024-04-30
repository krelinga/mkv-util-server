package main

import (
    "context"
    "errors"
    "fmt"

    "github.com/krelinga/mkv-util-server/pb"
)

func getChaptersSimple(ctx context.Context, path string) (*pb.SimpleChapters, error) {
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
