package main

import (
    "bytes"
    "context"
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"

    "connectrpc.com/connect"
    "github.com/google/go-cmp/cmp"
    "github.com/google/uuid"
    "github.com/krelinga/kgo/ktestcont"
    "google.golang.org/protobuf/testing/protocmp"
    "google.golang.org/protobuf/types/known/durationpb"

    pb "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video/mkv_util_server/v1"
    pbconnect "buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video/mkv_util_server/v1/mkv_util_serverv1connect"
)

type testContainer struct {
    containerId string
}

func newTestContainer() testContainer {
    return testContainer{
        containerId: fmt.Sprintf("mkv-util-server-docker-test-%s", uuid.NewString()),
    }
}

func captureOutput(cmd *exec.Cmd) *bytes.Buffer {
    cmdOutput := &bytes.Buffer{}
    cmd.Stdout = cmdOutput
    cmd.Stderr = cmdOutput
    return cmdOutput
}

func unsafeDuration(s string) time.Duration {
    d, err := time.ParseDuration(s)
    if err != nil {
        panic(err)
    }
    return d
}

func unsafeDurationPb(s string) *durationpb.Duration {
    return durationpb.New(unsafeDuration(s))
}

func readDuration(t *testing.T, mkvPath string, c pbconnect.MkvUtilServiceClient) time.Duration {
    t.Helper()
    req := connect.NewRequest(&pb.GetInfoRequest{
        InPath: mkvPath,
    })
    repl, err := c.GetInfo(context.Background(), req)
    if err != nil {
        t.Fatal(err)
    }
    return repl.Msg.Info.Duration.AsDuration()
}

func unsafeOutputPath(t *testing.T, suffix string) string {
    wd, err := os.Getwd()
    if err != nil {
        t.Fatalf("could not get working directory: %s", err)
    }
    localTestdataPath := filepath.Join(wd, "testdata")
    dockerTestDataPath := "/testdata"
    // localDir will be evaluated in this process.
    localDir := filepath.Join(localTestdataPath, "out", t.Name())
    if err := os.MkdirAll(localDir, 0755); err != nil {
        panic(err)
    }
    // returned path will be evaluated in the docker container.
    return filepath.Join(dockerTestDataPath, "out", t.Name(), suffix)
}

func (tc testContainer) Build(t *testing.T) {
    t.Helper()
    cmd := exec.Command("docker", "image", "build", "-t", tc.containerId, ".")
    cmdOutput := captureOutput(cmd)
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not build docker container: %s %s", err, cmdOutput)
    }
    t.Log("Finished building docker container.")
}

func (tc testContainer) Stop(t *testing.T) {
    t.Helper()
    cmd := exec.Command("docker", "container", "stop", tc.containerId)
    cmdOutput := captureOutput(cmd)
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not stop & delete docker container: %s %s", err, cmdOutput)
    }
    t.Log("Finished stopping & deleting docker container.")

    cmd = exec.Command("docker", "image", "rm", tc.containerId)
    cmdOutput = captureOutput(cmd)
    if err := cmd.Run(); err != nil {
        t.Fatalf("could not delete docker image: %s %s", err, cmdOutput)
    }
    t.Log("Finished deleting docker image.")
}

func (tc testContainer) Run(t *testing.T, shareDir string) {
    t.Helper()
    cmd := exec.Command("docker")
    wd, err := os.Getwd()
    if err != nil {
        t.Fatalf("could not get working directory: %s", err)
    }
    testdataPath := filepath.Join(wd, "testdata")
    mountCfg := fmt.Sprintf("type=bind,source=%s,target=/testdata", testdataPath)
    shareMountCfg := fmt.Sprintf("type=bind,source=%s,target=%s", shareDir, shareDir)
    userCfg := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
    args := []string{
        "run",
        "--rm",
        "-d",
        "--name", tc.containerId,
        "-p", "25002:25002",
        "--mount", mountCfg,
        "--mount", shareMountCfg,
        // This is needed so that generated files & directories have the correct ownership.
        "--user", userCfg,
        tc.containerId,
    }
    cmdOutput := captureOutput(cmd)
    cmd.Args = append(cmd.Args, args...)
    if err := cmd.Run(); err != nil {
        t.Fatalf("Could not run docker container: %s %s", err, cmdOutput)
    }
    t.Log("Started Docker container.")
}

func testGetFileSize(t *testing.T, c pbconnect.MkvUtilServiceClient) {
    req := connect.NewRequest(&pb.GetFileSizeRequest{
        Path: "/testdata/test.txt",
    })
    rep, err := c.GetFileSize(context.Background(), req)
    if err != nil {
        t.Errorf("Error calling GetFileSize(): %s", err)
        return
    }
    stat, err := os.Stat("testdata/test.txt")
    if err != nil {
        t.Errorf("Error stat'ing test file: %s", err)
        return
    }
    if rep.Msg.Size != stat.Size() {
        t.Errorf("size mismatch: rep.Size = %d, stat.Size() = %d", rep.Msg.Size, stat.Size())
    }
}

func testRunMkvToolNixCommand(t *testing.T, c pbconnect.MkvUtilServiceClient) {
    t.Run("File Exists", func(t *testing.T) {
        req := connect.NewRequest(&pb.RunMkvToolNixCommandRequest{
            Command: pb.RunMkvToolNixCommandRequest_COMMAND_MKVINFO,
            Args: []string{
                "/testdata/sample_640x360.mkv",
            },
        })
        resp, err := c.RunMkvToolNixCommand(context.Background(), req)
        if err != nil || len(resp.Msg.Stdout) == 0 {
            t.Errorf("Error calling mkvinfo: %s", err)
            if resp != nil {
                t.Errorf("Stdout:\n%s", resp.Msg.Stdout)
                t.Errorf("Stderr:\n%s", resp.Msg.Stderr)
            }
        }
    })
    t.Run("File does not exist", func(t *testing.T) {
        req := connect.NewRequest(&pb.RunMkvToolNixCommandRequest{
            Command: pb.RunMkvToolNixCommandRequest_COMMAND_MKVINFO,
            Args: []string{
                "/does/not/exist",
            },
        })
        resp, err := c.RunMkvToolNixCommand(context.Background(), req)
        if err == nil {
            t.Errorf("Stdout:\n%s", resp.Msg.Stdout)
            t.Errorf("Stderr:\n%s", resp.Msg.Stderr)
        }
    })
}

func readChapters(t *testing.T, p string, c pbconnect.MkvUtilServiceClient) *pb.SimpleChapters {
    t.Helper()
    req := connect.NewRequest(&pb.GetChaptersRequest{
        InPath: p,
        Format: pb.ChaptersFormat_CHAPTERS_FORMAT_SIMPLE,
    })
    resp, err := c.GetChapters(context.Background(), req)
    if err != nil {
        t.Errorf("Could not get chapters in %s: %s", p, err)
        return &pb.SimpleChapters{}
    }
    return resp.Msg.Chapters.Simple
}

func countChapters(t *testing.T, p string, c pbconnect.MkvUtilServiceClient) int {
    t.Helper()
    return len(readChapters(t, p, c).Chapters)
}

func testConcat(t *testing.T, c pbconnect.MkvUtilServiceClient) {
    t.Run("chapters_added_to_chapterless_file", func(t *testing.T) {
        inPath := "/testdata/sample_640x360.mkv"
        outPath:= unsafeOutputPath(t, "concat.mkv")
        req := connect.NewRequest(&pb.ConcatRequest{
            InputPaths: []string{inPath, inPath},
            OutputPath: outPath,
        })
        _, err := c.Concat(context.Background(), req)
        if err != nil {
            t.Error(err)
            return
        }

        d := readDuration(t, outPath, c)
        expD := unsafeDuration("26s692ms")
        if d != expD {
            t.Error(d)
        }

        if cnt := countChapters(t, outPath, c); cnt != 2 {
            t.Errorf("Expected 2 chapters in output, saw %d", cnt)
        }
    })
    t.Run("existing_chapters_preserved", func(t *testing.T) {
        outPath:= unsafeOutputPath(t, "concat.mkv")
        req := connect.NewRequest(&pb.ConcatRequest{
            InputPaths: []string{
                "/testdata/3_chapters.mkv",
                "/testdata/4_chapters.mkv",
            },
            OutputPath: outPath,
        })
        _, err := c.Concat(context.Background(), req)
        if err != nil {
            t.Error(err)
            return
        }

        actualChapters := readChapters(t, outPath, c)
        // TODO: Ideally we could test that each of these chapters are some
        // integer multiple of the sample file length, but it seems that due
        // to some accident of rounting somewhere, we get off by a millisecond
        // or two.  Not worth worrying about now, but would be good to track
        // down and fix eventually.
        expectedChaptres := &pb.SimpleChapters{
            Chapters: []*pb.SimpleChapters_Chapter{
                {
                    Number: 1,
                    Name: "Chapter 01",
                    Offset: durationpb.New(0),
                },
                {
                    Number: 2,
                    Name: "Chapter 02",
                    Offset: unsafeDurationPb("13.346s"),
                },
                {
                    Number: 3,
                    Name: "Chapter 03",
                    Offset: unsafeDurationPb("26.693s"),
                },
                {
                    Number: 4,
                    Name: "Chapter 04",
                    Offset: unsafeDurationPb("40.039s"),
                },
                {
                    Number: 5,
                    Name: "Chapter 05",
                    Offset: unsafeDurationPb("53.385s"),
                },
                {
                    Number: 6,
                    Name: "Chapter 06",
                    Offset: unsafeDurationPb("66.732s"),
                },
                {
                    Number: 7,
                    Name: "Chapter 07",
                    Offset: unsafeDurationPb("80.078s"),
                },
            },
        }
        if !cmp.Equal(expectedChaptres, actualChapters, protocmp.Transform()) {
            t.Errorf(cmp.Diff(expectedChaptres, actualChapters, protocmp.Transform()))
        }
    })
    t.Run("no_input_files_given", func(t *testing.T) {
        outPath:= unsafeOutputPath(t, "concat.mkv")
        req := connect.NewRequest(&pb.ConcatRequest{
            OutputPath: outPath,
        })
        _, err := c.Concat(context.Background(), req)
        if err == nil {
            t.Error("Expected an error.")
        }
    })
}

func testGetChapters(t *testing.T, c pbconnect.MkvUtilServiceClient) {
    req := connect.NewRequest(&pb.GetChaptersRequest{
        InPath: "/testdata/sample_640x360.mkv",
        Format: pb.ChaptersFormat_CHAPTERS_FORMAT_SIMPLE,
    })
    resp, err := c.GetChapters(context.Background(), req)
    if err != nil {
        t.Error(err)
        return
    }
    expected := &pb.GetChaptersResponse {
        Chapters: &pb.Chapters{
            Format: pb.ChaptersFormat_CHAPTERS_FORMAT_SIMPLE,
        },
    }
    if !cmp.Equal(expected, resp.Msg, protocmp.Transform()) {
        t.Error(cmp.Diff(expected, resp.Msg, protocmp.Transform()))
    }
}

func testGetInfo(t *testing.T, c pbconnect.MkvUtilServiceClient) {
    req := connect.NewRequest(&pb.GetInfoRequest{
        InPath: "/testdata/sample_640x360.mkv",
    })
    resp, err := c.GetInfo(context.Background(), req)
    if err != nil {
        t.Error(err)
        return
    }
    exp := &pb.GetInfoResponse{
        Info: &pb.Info{
            Duration: unsafeDurationPb("13s346ms"),
        },
    }
    if !cmp.Equal(exp, resp.Msg, protocmp.Transform()) {
        t.Error(cmp.Diff(exp, resp.Msg, protocmp.Transform()))
    }
}

func testSplit(t *testing.T, c pbconnect.MkvUtilServiceClient) {
    t.Run("empty_in_path_causes_error", func(t *testing.T) {
        req := connect.NewRequest(&pb.SplitRequest{})
        _, err := c.Split(context.Background(), req)
        if err == nil {
            t.Error("Expected an error.")
        }
    })
    t.Run("start_too_high_causes_error", func(t *testing.T) {
        req := connect.NewRequest(&pb.SplitRequest{
            InPath: "/testdata/3_chapters.mkv",
            ByChapters: []*pb.SplitRequest_ByChapters{
                {
                    Start: 10,
                    OutPath: unsafeOutputPath(t, "split.mkv"),
                },
            },
        })
        _, err := c.Split(context.Background(), req)
        if err == nil {
            t.Error("Expected an error.")
        }
    })
    t.Run("limit_too_high_causes_error", func(t *testing.T) {
        req := connect.NewRequest(&pb.SplitRequest{
            InPath: "/testdata/3_chapters.mkv",
            ByChapters: []*pb.SplitRequest_ByChapters{
                {
                    Limit: 10,
                    OutPath: unsafeOutputPath(t, "split.mkv"),
                },
            },
        })
        _, err := c.Split(context.Background(), req)
        if err == nil {
            t.Error("Expected an error.")
        }
    })
    t.Run("range_in_middle", func(t *testing.T) {
        outPath := unsafeOutputPath(t, "split.mkv")
        req := connect.NewRequest(&pb.SplitRequest{
            InPath: "/testdata/4_chapters.mkv",
            ByChapters: []*pb.SplitRequest_ByChapters{
                {
                    Start: 2,
                    Limit: 4,
                    OutPath: outPath,
                },
            },
        })
        _, err := c.Split(context.Background(), req)
        if err != nil {
            t.Error(err)
            return
        }
        actualD := readDuration(t, outPath, c)
        expectedD := unsafeDuration("26.693s")
        if actualD != expectedD {
            t.Error(actualD.String())
        }

        expectedChaps := &pb.SimpleChapters{
            Chapters: []*pb.SimpleChapters_Chapter{
                {
                    Number: 1,
                    Name: "Chapter 01",
                    Offset: unsafeDurationPb("0s"),
                },
                {
                    Number: 2,
                    Name: "Chapter 02",
                    Offset: unsafeDurationPb("13.347s"),
                },
            },
        }
        actualChaps := readChapters(t, outPath, c)
        if !cmp.Equal(expectedChaps, actualChaps, protocmp.Transform()) {
            t.Error(cmp.Diff(expectedChaps, actualChaps, protocmp.Transform()))
        }
    })
}

func TestDocker(t *testing.T) {
    if testing.Short() {
        t.Skip()
        return
    }
    t.Parallel()
    env, err := ktestcont.NewEnv(t)
    if err != nil {
        t.Fatal(err)
    }
    defer env.Cleanup()
    if err := os.RemoveAll("testdata/out"); err != nil {
        t.Fatalf("Could not remove existing test output directory: %s", err)
    }
    tc := newTestContainer()
    tc.Build(t)
    tc.Run(t, env.SharedDir())
    defer tc.Stop(t)

    // Create a stub to the test server.
    const target = "http://docker-daemon:25002"
    client := pbconnect.NewMkvUtilServiceClient(http.DefaultClient, target)

    t.Run("testGetFileSize", func(t *testing.T) {
        testGetFileSize(t, client)
    })
    t.Run("testRunMkvToolNixCommand", func(t *testing.T) {
        testRunMkvToolNixCommand(t, client)
    })
    t.Run("testConcat", func(t *testing.T) {
        testConcat(t, client)
    })
    t.Run("testGetInfo", func(t *testing.T) {
        testGetInfo(t, client)
    })
    t.Run("testSplit", func(t *testing.T) {
        testSplit(t, client)
    })
}
