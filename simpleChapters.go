package main

import (
    "bufio"
    //"errors"
    "fmt"
    "io"
    "regexp"
    //"strconv"
    "strings"
    //"time"

    "github.com/krelinga/mkv-util-server/pb"
)

var (
    simpleChapterRe = regexp.MustCompile(`^CHAPTER(\d+)=(\d+:\d+:\d+\.\d+)
CHAPTER(\d+)NAME=(.+)`)
)

//func parseChapterStartTime(t string) (time.Duration, error) {
//
//}

func parseSimpleChapters(r io.Reader) (*pb.SimpleChapters, error) {
//    extractInt := func(x string) (int, error) {
//        if len(x) == 0 {
//            return nil, errors.New("no match")
//        }
//        y, err := strconv.Atoi(x)
//        if err != nil {
//            return nil, fmt.Errorf("could not convert '%s' to an int: %e", x, err)
//        }
//        return y
//    }
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
            if len(matches) != 5 {
                return nil, fmt.Errorf("wrong number of matches: %d", len(matches))
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
