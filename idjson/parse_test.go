package idjson

import (
    "io"
    "os"
    "path/filepath"
    "runtime"
    "testing"
    "time"

    "github.com/google/go-cmp/cmp"
)

func unsafeOpenTestData() io.ReadCloser {
    _, thisPath, _, _ := runtime.Caller(0)
    p := filepath.Join(filepath.Dir(thisPath), "../testdata/id.json")
    f, err := os.Open(p)
    if err != nil {
        panic(err)
    }
    return f
}

func TestParse(t *testing.T) {
    t.Parallel()
    f := unsafeOpenTestData()
    defer f.Close()

    out, err := Parse(f)
    if err != nil {
        t.Error(err)
    }
    expected := &MkvMerge{
        Tracks: []*Track{
            {
                Type: "video",
                Properties: &TrackProperties{
                    TagDuration: "00:00:13.346000000",
                },
            },
        },
    }
    if !cmp.Equal(expected, out) {
        t.Error(cmp.Diff(expected, out))
    }
}

func TestParseTagDuration(t *testing.T) {
    t.Parallel()
    cases := []struct{
        Raw string
        Dur string
        Err error
    } {
        {
            Raw: "01:02:13.346000000",
            Dur: "1h2m13s346ms",
        },
        {
            Raw: "01:02:13.34600000",
            Err: ParseTagDurationWrongFormat,
        },
        {
            Raw: "1:2:13.346000000",
            Err: ParseTagDurationWrongFormat,
        },
    }
    for _, c := range cases {
        c := c
        t.Run(c.Raw, func(t *testing.T) {
            in := TrackProperties{
                TagDuration: c.Raw,
            }
            act, err := in.ParseTagDuration()
            if c.Err != nil && c.Err != err {
                t.Error(err)
            }
            if c.Dur != "" {
                exp, badErr := time.ParseDuration(c.Dur)
                if badErr != nil {
                    t.Fatal(badErr)
                }
                if exp != act {
                    t.Error(act)
                }
            }
        })
    }
}
