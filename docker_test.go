package main

import (
    "bytes"
    "fmt"
    "testing"
    "os/exec"

    "github.com/google/uuid"
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
        tc.containerId,
    }
    cmdOutput := captureOutput(cmd)
    cmd.Args = append(cmd.Args, args...)
    if err := cmd.Run(); err != nil {
        t.Fatalf("Could not run docker container: %s %s", err, cmdOutput)
    }
    t.Log("Started Docker container.")
}

func testGetFileSize(t *testing.T) {
    t.Parallel()
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

    t.Run("testGetFileSize", testGetFileSize)
}
