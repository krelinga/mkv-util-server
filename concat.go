package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "sync"
    "time"

    "google.golang.org/protobuf/types/known/durationpb"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
)

func concat(ctx context.Context, r *pb.ConcatRequest) (*pb.ConcatResponse, error) {
    type input struct {
        Path string
        Duration time.Duration
        Chapters *pb.SimpleChapters
    }
    inputs := []*input{}
    wg := sync.WaitGroup{}
    ctx, cancel := context.WithCancelCause(ctx)
    for _, p := range r.InputPaths {
        p := p
        i := &input{
            Path: p,
        }
        inputs = append(inputs, i)
        wg.Add(1)
        go func() {
            defer wg.Done()
            req := &pb.GetInfoRequest{
                InPath: i.Path,
            }
            resp, err := getInfo(ctx, req)
            if err != nil {
                cancel(err)
                return
            }
            i.Duration = resp.Info.Duration.AsDuration()
        }()
        wg.Add(1)
        go func() {
            defer wg.Done()
            req := &pb.GetChaptersRequest{
                InPath: i.Path,
                Format: pb.ChaptersFormat_CHAPTERS_FORMAT_SIMPLE,
            }
            resp, err := getChapters(ctx, req)
            if err != nil {
                cancel(err)
                return
            }
            i.Chapters = resp.Chapters.Simple
        }()
    }
    wg.Wait()
    // See if anything went wrong during the info gathering.
    if err := context.Cause(ctx); err != nil {
        return nil, err
    }


    // Make sure there's a zero-offset chapter at the beginning of every input.
    for _, i := range inputs {
        cs := &i.Chapters.Chapters
        if len(*cs) == 0 || (*cs)[0].Offset.AsDuration() != time.Duration(0) {
            zeroC := &pb.SimpleChapters_Chapter{
                Offset: durationpb.New(0),
            }
            *cs = append([]*pb.SimpleChapters_Chapter{zeroC,}, (*cs)...)
        }
    }

    // Build final chapters, updating the offset of each chapter by the duration
    // of the videos that came before it.
    chapterName := func(i int32) string {
        return fmt.Sprintf("Chapter %02d", i)
    }
    chapterNum := int32(1)
    chaps := &pb.SimpleChapters{}
    cumD := time.Duration(0)
    for _, i := range inputs {
        for _, ic := range i.Chapters.Chapters {
            chaps.Chapters = append(chaps.Chapters, &pb.SimpleChapters_Chapter{
                Number: chapterNum,
                Name: chapterName(chapterNum),
                Offset: durationpb.New(cumD + ic.Offset.AsDuration()),
            })
            chapterNum++
        }
        cumD += i.Duration
    }

    // Store the chapters file out on disk
    tmpDir, err := os.MkdirTemp("", "")
    if err != nil {
        return nil, fmt.Errorf("Could not create temporary dir: %w", err)
    }
    defer func() {
        if err := os.RemoveAll(tmpDir); err != nil {
            log.Printf("Could not remove temporary dir: %s", err)
        }
    }()
    chPath := filepath.Join(tmpDir, "chapters")
    chFile, err := os.Create(chPath)
    if err != nil {
        return nil, fmt.Errorf("Could not open %s for writing: %w", chPath, err)
    }
    if err := writeSimpleChapters(chFile, chaps); err != nil {
        chFile.Close()
        return nil, fmt.Errorf("Could not write chapters to file: %w", err)
    }
    if err := chFile.Close(); err != nil {
        return nil, fmt.Errorf("Could not close chapters file: %w", err)
    }

    args := []string{
        "-o", r.OutputPath,
        "--chapters", chPath,
    }
    for i, p := range r.InputPaths {
        args = append(args, "--no-chapters")
        if i > 0 {
            args = append(args, "+")
        }
        args = append(args, p)
    }
    cmd := exec.CommandContext(ctx, "mkvmerge", args...)
    cmd.Stdout = log.Default().Writer()
    cmd.Stderr = log.Default().Writer()
    if err := cmd.Run(); err != nil {
        return nil, err
    }
    return &pb.ConcatResponse{}, nil
}
