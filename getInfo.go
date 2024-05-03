package main

import (
    "context"
    "errors"

    "github.com/krelinga/mkv-util-server/pb"
)

func getInfo(ctx context.Context, r *pb.GetInfoRequest) (*pb.GetInfoReply, error) {
    return nil, errors.New("Unimplemented")
}
