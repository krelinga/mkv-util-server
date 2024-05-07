package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/krelinga/mkv-util-server/pb"
)

func split(ctx context.Context, r *pb.SplitRequest) (*pb.SplitReply, error) {
    chaps, err := getChapters(ctx, &pb.GetChaptersRequest{
        Format: pb.ChaptersFormat_CF_SIMPLE,
        InPath: r.InPath,
    })
    if err != nil {
        return nil, fmt.Errorf("Could not get chapters: %e", err)
    }
    chapsIndex := map[int32]*pb.SimpleChapters_Chapter{}
    for _, c := range chaps.Chapters.Simple.Chapters {
        chapsIndex[c.Number] = c
    }
    type output struct {
        Path string
        Begin time.Duration
        End time.Duration
        Chapters *pb.SimpleChapters
    }
    outputs := []*output{}
    for _, o := range r.ByChapters {
        newOut := &output{
            Path: o.OutPath,
            Chapters: &pb.SimpleChapters{},
        }
        beginN := int32(0)
        if o.Start != 0 {
            beginC, found := chapsIndex[o.Start]
            beginN = beginC.Number
            if !found {
                return nil, fmt.Errorf("Invalid chapter %d", o.Start)
            }
            newOut.Begin = beginC.Offset.AsDuration()
        }
        endN := int32(math.MaxInt32)
        if o.Limit != 0 {
            endC, found := chapsIndex[o.Limit]
            endN = endC.Number
            if !found {
                return nil, fmt.Errorf("Invalid chapter %d", o.Limit)
            }
            newOut.End = endC.Offset.AsDuration()
        }
        chapNumber := int32(1)
        chapName := func(i int32) string {
            return fmt.Sprintf("Chapter %02d", i)
        }
        for _, c := range chaps.Chapters.Simple.Chapters {
            if c.Number >= beginN && c.Number < endN {
                newOut.Chapters.Chapters = append(newOut.Chapters.Chapters, &pb.SimpleChapters_Chapter{
                    Number: chapNumber,
                    Name: chapName(chapNumber),
                    Offset: c.Offset,
                })
                chapNumber++
            }
        }
        outputs = append(outputs, newOut)
    }

    ctx, cancel := context.WithCancelCause(ctx)
    wg := sync.WaitGroup{}
    for _, o := range outputs {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // Write chapters to a temporary file.
            tmpDir, err := os.MkdirTemp("", "")
            if err != nil {
                cancel(fmt.Errorf("Could not create temporary dir: %e", err))
                return
            }
            defer func() {
                if err := os.RemoveAll(tmpDir); err != nil {
                    log.Printf("Could not remove temporary dir: %s", err)
                }
            }()
            chPath := filepath.Join(tmpDir, "chapters")
            chFile, err := os.Create(chPath)
            if err != nil {
                cancel(fmt.Errorf("Could not open %s for writing: %e", chPath, err))
                return
            }
            if err := writeSimpleChapters(chFile, o.Chapters); err != nil {
                chFile.Close()
                cancel(fmt.Errorf("Could not write chapters to file: %e", err))
                return
            }
            if err := chFile.Close(); err != nil {
                cancel(fmt.Errorf("Could not close chapters file: %e", err))
                return
            }

            sb := strings.Builder{}
            sb.WriteString("parts:")
            if o.Begin != time.Duration(0) {
                sb.WriteString(o.Begin.String())
            }
            sb.WriteString("-")
            if o.End != time.Duration(0) {
                sb.WriteString(o.End.String())
            }
            args := []string{
                "-o", o.Path,
                "--chapters", chPath,
                "--split",
                sb.String(),
                "--no-chapters",
                r.InPath,
            }
            cmd := exec.CommandContext(ctx, "mkvmerge", args...)
            cmd.Stderr = log.Default().Writer()
            cmd.Stdout = log.Default().Writer()
            if err := cmd.Run(); err != nil {
                cancel(err)
                return
            }
        }()
    }
    wg.Wait()
    if context.Cause(ctx) != nil {
        return nil, context.Cause(ctx)
    }
    return &pb.SplitReply{}, nil
}
