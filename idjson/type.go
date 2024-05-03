package idjson

type MkvMerge struct {
    Tracks []*Track
}

type Track struct {
    Type string
    Properties *TrackProperties
}

type TrackProperties struct {
    TagDuration string `json:"tag_duration"`
}
