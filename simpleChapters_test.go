package main

import (
    "strings"
    "testing"
    "time"

    "github.com/google/go-cmp/cmp"
    "google.golang.org/protobuf/testing/protocmp"
    "google.golang.org/protobuf/types/known/durationpb"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
)

const (
    exampleSimpleChapters = `CHAPTER01=00:00:00.000
CHAPTER01NAME=Intro
CHAPTER02=00:02:30.000
CHAPTER02NAME=Baby prepares to rock
CHAPTER03=01:02:42.300
CHAPTER03NAME=Baby rocks the house`
)

// Converts to a duration, panics on error
func unsafeProtoDuration(s string) *durationpb.Duration {
    d, err := time.ParseDuration(s)
    if err != nil {
        panic(err)
    }
    return durationpb.New(d)
}

func TestParseSimpleChapters(t *testing.T) {
    t.Parallel()
    c, err := parseSimpleChapters(strings.NewReader(exampleSimpleChapters))
    if err != nil {
        t.Error(err)
        return
    }
    expected := &pb.SimpleChapters{
        Chapters: []*pb.SimpleChapters_Chapter{
            {
                Number: 1,
                Name: "Intro",
                Offset: unsafeProtoDuration("0"),
            },
            {
                Number: 2,
                Name: "Baby prepares to rock",
                Offset: unsafeProtoDuration("2m30s"),
            },
            {
                Number: 3,
                Name: "Baby rocks the house",
                Offset: unsafeProtoDuration("1h2m42s300ms"),
            },
        },
    }
    // TODO: diff ouptut for the `Offset` field is not very useful.
    if !cmp.Equal(expected, c, protocmp.Transform()) {
        t.Error(cmp.Diff(expected, c, protocmp.Transform()))
    }

    // TODO: test cases to exercise all of the error detection logic?
}

func TestWriteSimpleChapters(t *testing.T) {
    t.Parallel()
    in := &pb.SimpleChapters{
        Chapters: []*pb.SimpleChapters_Chapter{
            {
                Number: 1,
                Name: "taters",
                Offset: unsafeProtoDuration("0s"),
            },
            {
                Number: 2,
                Name: "pie",
                Offset: unsafeProtoDuration("4h10m31s15ms"),
            },
        },
    }
    expected := `CHAPTER01=00:00:00.000
CHAPTER01NAME=taters
CHAPTER02=04:10:31.015
CHAPTER02NAME=pie
`
    sb := strings.Builder{}
    if err := writeSimpleChapters(&sb, in); err != nil {
        t.Error(err)
        return
    }
    actual := sb.String()
    if !cmp.Equal(expected, actual) {
        t.Error(cmp.Diff(expected, actual))
    }
}
