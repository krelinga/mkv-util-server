package idjson

import (
    "fmt"
    "io"
    "encoding/json"
    "errors"
)

func Parse(r io.Reader) (*MkvMerge, error) {
    d := json.NewDecoder(r)
    m := &MkvMerge{}
    if err := d.Decode(m); err != nil {
        return nil, fmt.Errorf("Could not decode: %e", err)
    }
    if d.More() {
        return nil, errors.New("More than one element in input")
    }
    return m, nil
}
