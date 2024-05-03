package main

import (
    "strings"
    "testing"
    "time"

    "github.com/google/go-cmp/cmp"
    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/protobuf/testing/protocmp"
    "google.golang.org/protobuf/types/known/durationpb"
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
