package main

import (
    "bytes"
    "context"
    "fmt"
    "testing"
    "os/exec"

    "github.com/google/uuid"
    "github.com/krelinga/mkv-util-server/pb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
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
    args := []string{
        "run",
        "--rm",
        "-d",
        "--name", tc.containerId,
        "-p", "0.0.0.0:25002:25002",
        tc.containerId,
    }
    cmdOutput := captureOutput(cmd)
    cmd.Args = append(cmd.Args, args...)
    if err := cmd.Run(); err != nil {
        t.Fatalf("Could not run docker container: %s %s", err, cmdOutput)
    }
    t.Log("Started Docker container.")
}

func testGetFileSize(t *testing.T, c pb.MkvUtilsClient) {
    req := &pb.GetFileSizeRequest{}
    rep, err := c.GetFileSize(context.Background(), req)
    if err != nil {
        t.Error(err)
        return
    }
    if rep.Size != -1 {
        t.Error(rep.Size)
    }
}

func TestDocker(t *testing.T) {
    if testing.Short() {
        t.Skip()
        return
    }
    t.Parallel()
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
    client := pb.NewMkvUtilsClient(conn)

    t.Run("testGetFileSize", func(t *testing.T) {
        testGetFileSize(t, client)
    })
}
