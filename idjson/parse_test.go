package idjson

import (
    "io"
    "os"
    "path/filepath"
    "runtime"
    "testing"

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
