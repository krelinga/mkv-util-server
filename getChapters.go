package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"

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
    chPath := filepath.Join(tmpDir, "chapters")
    cmd := exec.CommandContext(ctx, "mkvextract", path, "chapters", "-s", chPath)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("Error running mkvextract: %e", err)
    }
    chFile, err := os.Open(chPath)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            // special case: if no chapters were found then no output file is
            // produced.
            return &pb.SimpleChapters{}, nil
        }
        return nil, fmt.Errorf("Could not open chapter file: %e", err)
    }
    defer chFile.Close()
    return parseSimpleChapters(chFile)
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
