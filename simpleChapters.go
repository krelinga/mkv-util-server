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

    "google.golang.org/protobuf/types/known/durationpb"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
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
                return nil, fmt.Errorf("could not parse chapter start time: %w", err)
            }
            num1, err := strconv.Atoi(matches[1])
            if err != nil {
                return nil, fmt.Errorf("Could not extract 1st chapter number: %w", err)
            }
            num2, err := strconv.Atoi(matches[3])
            if err != nil {
                return nil, fmt.Errorf("Could not extract 2nd chapter number: %w", err)
            }
            if num1 != num2 {
                return nil, fmt.Errorf("Mismatched chapter numbers: %d vs %d", num1, num2)
            }
            name := matches[4]

            c := &pb.SimpleChapters_Chapter{
                Number: int32(num1),
                Name: name,
                Offset: durationpb.New(offset),
            }
            chapters.Chapters = append(chapters.Chapters, c)
            part = ""
        }
        first = !first
    }
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("could not scan simple chapters file: %w", err)
    }
    return chapters, nil
}

func setSimpleChaptersDurations(ch *pb.SimpleChapters, overallDuration time.Duration){
    for i := 0; i < len(ch.Chapters); i++ {
        nextOffset := func() time.Duration {
            if i < len(ch.Chapters) - 1 {
                return ch.Chapters[i + 1].Offset.AsDuration()
            } else {
                return overallDuration
            }
        }()
        ch.Chapters[i].Duration = durationpb.New(nextOffset - ch.Chapters[i].Offset.AsDuration())
    }
}

func writeSimpleChapters(w io.Writer, ch *pb.SimpleChapters) error {
    h := func(d time.Duration) int {
        return int(d.Truncate(time.Hour) / time.Hour)
    }
    m := func(d time.Duration) int {
        return int((d.Truncate(time.Minute) % time.Hour) / time.Minute)
    }
    s := func(d time.Duration) int {
        return int((d.Truncate(time.Second) % time.Minute) / time.Second)
    }
    ms := func(d time.Duration) int {
        return int((d.Truncate(time.Millisecond) % time.Second) / time.Millisecond)
    }
    for _, c := range ch.Chapters {
        d := c.Offset.AsDuration()
        sd := fmt.Sprintf("%02d:%02d:%02d.%03d", h(d), m(d), s(d), ms(d))
        _, err := fmt.Fprintf(w, "CHAPTER%02d=%s\nCHAPTER%02dNAME=%s\n", c.Number, sd, c.Number, c.Name)
        if err != nil {
            return err
        }
    }
    return nil
}
