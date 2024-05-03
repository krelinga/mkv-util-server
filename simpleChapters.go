package main

import (
    "bufio"
    "fmt"
    "io"
    "regexp"
    "strings"

    "github.com/krelinga/mkv-util-server/pb"
)

var (
    simpleChapterRe = regexp.MustCompile(`^CHAPTER(\d+)=(\d)+:(\d+):(\d+)\.(\d+)
CHAPTER(\d+)NAME=(.+)`)
)

func parseSimpleChapters(r io.Reader) (*pb.SimpleChapters, error) {
    scanner := bufio.NewScanner(r)
    first := true
    var part string
    chapters := &pb.SimpleChapters{}
    for scanner.Scan() {
        if first {
            part = scanner.Text()
        } else {
            part = strings.Join([]string{part, scanner.Text()}, "\n")
            matches := simpleChapterRe.FindStringSubmatch(part)
            for i, match := range matches {
                if len(match) == 0 {
                    return nil, fmt.Errorf("Could not match part %d", i)
                }
            }
            chapters.Chapters = append(chapters.Chapters, &pb.SimpleChapters_Chapter{})
            part = ""
        }
        first = !first
    }
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("could not scan simple chapters file: %e", err)
    }
    return chapters, nil
}