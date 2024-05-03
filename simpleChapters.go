package main

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/protobuf/types/known/durationpb"
)

var (
    simpleChapterRe = regexp.MustCompile(`^CHAPTER(\d+)=(\d+:\d+:\d+\.\d+)
CHAPTER(\d+)NAME=(.+)`)
)

func parseChapterStartTime(t string) (time.Duration, error) {
    if len(t) == 0 {
        return 0, errors.New("no match")
    }
    t = strings.Replace(t, ":", "h", 1)
    t = strings.Replace(t, ":", "m", 1)
    t = strings.Replace(t, ".", "s", 1)
    t += "ms"
    return time.ParseDuration(t)
}

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
            offset, err := parseChapterStartTime(matches[2])
            if err != nil {
                return nil, fmt.Errorf("could not parse chapter start time: %e", err)
            }
            num1, err := strconv.Atoi(matches[1])
            if err != nil {
                return nil, fmt.Errorf("Could not extract 1st chapter number: %e", err)
            }
            num2, err := strconv.Atoi(matches[3])
            if err != nil {
                return nil, fmt.Errorf("Could not extract 2nd chapter number: %e", err)
            }
            if num1 != num2 {
                return nil, fmt.Errorf("Mismatched chapter numbers: %d vs %d", num1, num2)
            }

            c := &pb.SimpleChapters_Chapter{
                Number: int32(num1),
                Offset: durationpb.New(offset),
            }
            chapters.Chapters = append(chapters.Chapters, c)
            part = ""
        }
        first = !first
    }
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("could not scan simple chapters file: %e", err)
    }
    return chapters, nil
}
