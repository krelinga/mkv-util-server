package main

import (
    "strings"
    "testing"
)

const (
    exampleSimpleChapters = `CHAPTER01=00:00:00.000
CHAPTER01NAME=Intro
CHAPTER02=00:02:30.000
CHAPTER02NAME=Baby prepares to rock
CHAPTER03=00:02:42.300
CHAPTER03NAME=Baby rocks the house`
)

func TestParseSimpleChapters(t *testing.T) {
    c, err := parseSimpleChapters(strings.NewReader(exampleSimpleChapters))
    if err != nil {
        t.Error(err)
        return
    }
    if len(c.Chapters) != 3 {
        t.Error(len(c.Chapters))
    }
}
