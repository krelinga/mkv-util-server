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

    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/protobuf/types/known/durationpb"
)

func concat(ctx context.Context, r *pb.ConcatRequest) (*pb.ConcatReply, error) {
    type input struct {
        Path string
        Duration time.Duration
    }
    inputs := []*input{}
    wg := sync.WaitGroup{}
    ctx, cancel := context.WithCancelCause(ctx)
    for _, p := range r.InputPaths {
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
    }
    wg.Wait()
    // See if anything went wrong during the info gathering.
    if err := context.Cause(ctx); err != nil {
        return nil, err
    }

    chapterName := func(i int) string {
        return fmt.Sprintf("Chapter %02d", i)
    }
    // First chapter always begins at the start of the file.
    chaps := &pb.SimpleChapters{
        Chapters: []*pb.SimpleChapters_Chapter{
            {
                Number: 1,
                Name: chapterName(1),
                Offset: durationpb.New(0),
            },
        },
    }
    // Subsequent chapters begin at the end of every previous input.
    for i := 1; i < len(inputs); i++ {
        chaps.Chapters = append(chaps.Chapters, &pb.SimpleChapters_Chapter{
            Number: int32(i + 1),
            Name: chapterName(i + 1),
            Offset: durationpb.New(inputs[i - 1].Duration),
        })
    }

    // Store the chapters file out on disk
    tmpDir, err := os.MkdirTemp("", "")
    if err != nil {
        return nil, fmt.Errorf("Could not create temporary dir: %e", err)
    }
    defer func() {
        if err := os.RemoveAll(tmpDir); err != nil {
            log.Printf("Could not remove temporary dir: %s", err)
        }
    }()
    chPath := filepath.Join(tmpDir, "chapters")
    chFile, err := os.Create(chPath)
    if err != nil {
        return nil, fmt.Errorf("Could not open %s for writing: %e", chPath, err)
    }
    if err := writeSimpleChapters(chFile, chaps); err != nil {
        chFile.Close()
        return nil, fmt.Errorf("Could not write chapters to file: %e", err)
    }
    if err := chFile.Close(); err != nil {
        return nil, fmt.Errorf("Could not close chapters file: %e", err)
    }

    args := []string{
        "-o", r.OutputPath,
        "--chapters", chPath,
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
