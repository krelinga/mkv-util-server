package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
)

func getChaptersSimple(ctx context.Context, path string) (*pb.SimpleChapters, error) {
    tmpDir, err := os.MkdirTemp("", "")
    if err != nil {
        return nil, fmt.Errorf("Could not create temporary directory: %w", err)
    }
    defer func() {
        if err := os.RemoveAll(tmpDir); err != nil {
            log.Printf("could not remove temp dir %s: %s", tmpDir, err)
        }
    }()
    chPath := filepath.Join(tmpDir, "chapters")
    cmd := exec.CommandContext(ctx, "mkvextract", path, "chapters", "-s", chPath)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("Error running mkvextract: %w", err)
    }
    chFile, err := os.Open(chPath)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            // special case: if no chapters were found then no output file is
            // produced.
            return &pb.SimpleChapters{}, nil
        }
        return nil, fmt.Errorf("Could not open chapter file: %w", err)
    }
    defer chFile.Close()
    parsed, err := parseSimpleChapters(chFile)
    if err != nil {
        return nil, err
    }

    overallDuration, err := func() (time.Duration, error) {
        req := &pb.GetInfoRequest{
            InPath: path,
        }
        info, err := getInfo(ctx, req)
        if err != nil {
            return 0, err
        }
        if err := info.Info.Duration.CheckValid(); err != nil {
            return 0, err
        }
        return info.Info.Duration.AsDuration(), nil
    }()

    setSimpleChaptersDurations(parsed, overallDuration)
    return parsed, nil
}

func getChapters(ctx context.Context, r *pb.GetChaptersRequest) (*pb.GetChaptersResponse, error) {
    resp := &pb.GetChaptersResponse {}
    switch r.Format {
    case pb.ChaptersFormat_CHAPTERS_FORMAT_SIMPLE:
        simple, err := getChaptersSimple(ctx, r.InPath)
        if err != nil {
            return nil, err
        }
        resp.Chapters = &pb.Chapters{
            Format: pb.ChaptersFormat_CHAPTERS_FORMAT_SIMPLE,
            Simple: simple,
        }
    default:
        return nil, fmt.Errorf("Unsupported format: %v", r.Format)
    }

    return resp, nil
}
