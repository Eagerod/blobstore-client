package main;

import (
    "os/exec"
    "testing"
)

import (
    "github.com/stretchr/testify/assert"
)

// NOTE: This package expects a fresh binary to have just been installed on the
// machine. This should be updated to either use source, or use the binary in
// the project's build directory if possible, but that may be more difficult 
// than needed.
func init() {
    if _, err := exec.LookPath("blob"); err != nil {
        panic("Failed to find executable to run system tests")
    }
}

func TestCommandLineInterfaceUpload(t *testing.T) {
    cmd := exec.Command("blob", "upload", "--filename", "clientlib/testing/Makefile", "--type", "text/plain", "--source", "../Makefile")
    output, err := cmd.CombinedOutput()
    if err != nil {
        assert.Failf(t, err.Error(), string(output))
    }

    assert.Nil(t, err)
}
