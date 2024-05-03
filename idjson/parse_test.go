package idjson

import (
    "io"
    "os"
    "path/filepath"
    "runtime"
    "testing"
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

    _, err := Parse(f)
    if err != nil {
        t.Error(err)
    }
}
