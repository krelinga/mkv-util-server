package main

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"

    "github.com/google/go-cmp/cmp"
    "github.com/google/uuid"
    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/protobuf/testing/protocmp"
    "google.golang.org/protobuf/types/known/durationpb"
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

func unsafeDurationPb(s string) *durationpb.Duration {
    d, err := time.ParseDuration(s)
    if err != nil {
        panic(err)
    }
    return durationpb.New(d)
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

func (tc testContainer) Run(t *testing.T) {
    t.Helper()
    cmd := exec.Command("docker")
    wd, err := os.Getwd()
    if err != nil {
        t.Fatalf("could not get working directory: %s", err)
    }
    testdataPath := filepath.Join(wd, "testdata")
    mountCfg := fmt.Sprintf("type=bind,source=%s,target=/testdata", testdataPath)
    userCfg := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
    args := []string{
        "run",
        "--rm",
        "-d",
        "--name", tc.containerId,
        "-p", "25002:25002",
        "--mount", mountCfg,
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

func testGetFileSize(t *testing.T, c pb.MkvUtilClient) {
    req := &pb.GetFileSizeRequest{
        Path: "/testdata/test.txt",
    }
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
    if rep.Size != stat.Size() {
        t.Errorf("size mismatch: rep.Size = %d, stat.Size() = %d", rep.Size, stat.Size())
    }
}

func testRunMkvToolNixCommand(t *testing.T, c pb.MkvUtilClient) {
    t.Run("File Exists", func(t *testing.T) {
        req := &pb.RunMkvToolNixCommandRequest{
            Command: pb.RunMkvToolNixCommandRequest_COMMAND_MKVINFO,
            Args: []string{
                "/testdata/sample_640x360.mkv",
            },
        }
        resp, err := c.RunMkvToolNixCommand(context.Background(), req)
        if err != nil || len(resp.Stdout) == 0 {
            t.Errorf("Error calling mkvinfo: %s", err)
            if resp != nil {
                t.Errorf("Stdout:\n%s", resp.Stdout)
                t.Errorf("Stderr:\n%s", resp.Stderr)
            }
        }
    })
    t.Run("File does not exist", func(t *testing.T) {
        req := &pb.RunMkvToolNixCommandRequest{
            Command: pb.RunMkvToolNixCommandRequest_COMMAND_MKVINFO,
            Args: []string{
                "/does/not/exist",
            },
        }
        resp, err := c.RunMkvToolNixCommand(context.Background(), req)
        if err == nil {
            t.Errorf("Stdout:\n%s", resp.Stdout)
            t.Errorf("Stderr:\n%s", resp.Stderr)
        }
    })
}

func testConcat(t *testing.T, c pb.MkvUtilClient) {
    req := &pb.ConcatRequest{
        InputPaths: []string{
            "/testdata/sample_640x360.mkv",
            "/testdata/sample_640x360.mkv",
        },
        OutputPath: "/testdata/out/concat.mkv",
    }
    _, err := c.Concat(context.Background(), req)
    if err != nil {
        t.Error(err)
    }
}

func testGetChapters(t *testing.T, c pb.MkvUtilClient) {
    req := &pb.GetChaptersRequest{
        InPath: "/testdata/sample_640x360.mkv",
        Format: pb.ChaptersFormat_CF_SIMPLE,
    }
    resp, err := c.GetChapters(context.Background(), req)
    if err != nil {
        t.Error(err)
        return
    }
    expected := &pb.GetChaptersReply {
        Chapters: &pb.Chapters{
            Format: pb.ChaptersFormat_CF_SIMPLE,
        },
    }
    if !cmp.Equal(expected, resp, protocmp.Transform()) {
        t.Error(cmp.Diff(expected, c, protocmp.Transform()))
    }
}

func testGetInfo(t *testing.T, c pb.MkvUtilClient) {
    req := &pb.GetInfoRequest{
        InPath: "/testdata/sample_640x360.mkv",
    }
    resp, err := c.GetInfo(context.Background(), req)
    if err != nil {
        t.Error(err)
        return
    }
    exp := &pb.GetInfoReply{
        Info: &pb.Info{
            Duration: unsafeDurationPb("13s346ms"),
        },
    }
    if !cmp.Equal(exp, resp, protocmp.Transform()) {
        t.Error(cmp.Diff(exp, resp, protocmp.Transform()))
    }
}

func TestDocker(t *testing.T) {
    if testing.Short() {
        t.Skip()
        return
    }
    t.Parallel()
    if err := os.RemoveAll("testdata/out"); err != nil {
        t.Fatalf("Could not remove existing test output directory: %s", err)
    }
    tc := newTestContainer()
    tc.Build(t)
    tc.Run(t)
    defer tc.Stop(t)

    // Create a stub to the test server.
    const target = "docker-daemon:25002"
    creds := grpc.WithTransportCredentials(insecure.NewCredentials())
    conn, err := grpc.DialContext(context.Background(), target, creds)
    if err != nil {
        t.Fatalf("Could not dial target: %s", err)
    }
    client := pb.NewMkvUtilClient(conn)

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
}
