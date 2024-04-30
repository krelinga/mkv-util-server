package main

import (
    "context"
    "errors"

    "github.com/krelinga/mkv-util-server/pb"
)

func getChapters(ctx context.Context, r *pb.GetChaptersRequest) (*pb.GetChaptersReply, error) {
    return nil, errors.New("Not implemented")
}
