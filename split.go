package main

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "sync"
    "time"

    "github.com/krelinga/mkv-util-server/pb"
)

func split(ctx context.Context, r *pb.SplitRequest) (*pb.SplitReply, error) {
    chaps, err := getChapters(ctx, &pb.GetChaptersRequest{
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
    }
    outputs := []*output{}
    for _, o := range r.ByChapters {
        newOut := &output{
            Path: o.OutPath,
        }
        if o.Start != 0 {
            beginC, found := chapsIndex[o.Start]
            if !found {
                return nil, fmt.Errorf("Invalid chapter %d", o.Start)
            }
            newOut.Begin = beginC.Offset.AsDuration()
        }
        if o.Limit != 0 {
            endC, found := chapsIndex[o.Limit]
            if !found {
                return nil, fmt.Errorf("Invalid chapter %d", o.Limit)
            }
            newOut.End = endC.Offset.AsDuration()
        }
        outputs = append(outputs, newOut)
    }

    ctx, cancel := context.WithCancelCause(ctx)
    wg := sync.WaitGroup{}
    for _, o := range outputs {
        wg.Add(1)
        go func() {
            defer wg.Done()
            sb := strings.Builder{}
            sb.WriteString("parts:")
            if o.Begin != time.Duration(0) {
                sb.WriteString(o.Begin.String())
            }
            sb.WriteString("-")
            if o.End != time.Duration(0) {
                sb.WriteString(o.End.String())
            }
            args := []string{"split", sb.String()}
            cmd := exec.CommandContext(ctx, "mkvmerge", args...)
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
