package idjson

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "regexp"
    "strings"
    "time"
)

func Parse(r io.Reader) (*MkvMerge, error) {
    d := json.NewDecoder(r)
    m := &MkvMerge{}
    if err := d.Decode(m); err != nil {
        return nil, fmt.Errorf("Could not decode: %w", err)
    }
    if d.More() {
        return nil, errors.New("More than one element in input")
    }
    return m, nil
}

var tagDurationRE = regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{9}`)
var (
    ParseTagDurationWrongFormat = errors.New("Wrong format, expected 00:00:00.000000000")
    ParseTagDurationFinalFormatBug = errors.New("Wrong final format for string passed to time.ParseDuration().  This indicates a bug in ParseTagDuration().")
)

func (tp *TrackProperties) ParseTagDuration() (time.Duration, error) {
    if !tagDurationRE.MatchString(tp.TagDuration) {
        return 0, ParseTagDurationWrongFormat
    }
    dFormat := tp.TagDuration
    dFormat = strings.Replace(dFormat, ":", "h", 1)
    dFormat = strings.Replace(dFormat, ":", "m", 1)
    dFormat = strings.Replace(dFormat, ".", "s", 1)
    dFormat += "ns"
    d, err := time.ParseDuration(dFormat)
    if err != nil {
        return 0, ParseTagDurationFinalFormatBug
    }
    return d, nil
}

func (cp *ContainerProperties) ParseDuration() (time.Duration, error) {
    return time.ParseDuration(fmt.Sprintf("%dns", cp.Duration))
}
