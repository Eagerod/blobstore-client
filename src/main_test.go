package main;

import (
    "bufio"
    "io/ioutil"
    "os"
    "os/exec"
    "testing"
)

import (
    "github.com/stretchr/testify/assert"
)

import (
    "blobapi"
)

// NOTE: This package expects a fresh binary to have just been installed on the
// machine. This should be updated to either use source, or use the binary in
// the project's build directory if possible, but that may be more difficult 
// than needed.
const testingAccessToken string = "ad4c3f2d4fb81f4118f837464b961eebda026d8c52a7cc967047cc3c2a3f6a43"
var blobCliHelpString *string

func init() {
    if _, err := exec.LookPath("blob"); err != nil {
        panic("Failed to find executable to run system tests")
    }

    cmd := exec.Command("blob", "-h")
    str, err := cmd.CombinedOutput()
    if err != nil {
        panic("Failed to get help text from blob binary")
    }

    stringVal := string(str)
    blobCliHelpString = &stringVal
}

func TestCommandLineInterfaceUpload(t *testing.T) {
    cmd := exec.Command("blob", "upload", "--filename", "clientlib/testing/Makefile", "--type", "text/plain", "--source", "../Makefile")

    cmd.Env = append(os.Environ(),
        "BLOBSTORE_READ_ACL=" + testingAccessToken,
        "BLOBSTORE_WRITE_ACL=" + testingAccessToken,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        assert.Failf(t, err.Error(), string(output))
    }

    assert.Nil(t, err)
    assert.Equal(t, "", string(output))

    api := blobapi.NewBlobStoreApiClient("https://aleem.haji.ca/blob", testingAccessToken, testingAccessToken)
    contents, err := api.GetFileContents("clientlib/testing/Makefile")
    assert.Nil(t, err)

    file, err := os.Open("../Makefile")
    assert.Nil(t, err)

    expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
    assert.Nil(t, err)

    assert.Equal(t, string(expectedBody), contents)
}

func TestCommandLineInterfaceUploadFails(t *testing.T) {
    cmd := exec.Command("blob", "upload", "--filename", "clientlib/testing/Makefile", "--type", "text/plain", "--source", "../Makefile")

    output, err := cmd.CombinedOutput()
    if err == nil {
        assert.Fail(t, "Expected a failure from download command")
    }

    expectedOutput := "Blobstore Upload Failed (403): \n" + *blobCliHelpString + "\n"
    assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceDownload(t *testing.T) {
    api := blobapi.NewBlobStoreApiClient("https://aleem.haji.ca/blob", testingAccessToken, testingAccessToken)
    api.UploadFile("clientlib/testing/Makefile", "../Makefile", "text/plain")

    cmd := exec.Command("blob", "download", "--filename", "clientlib/testing/Makefile", "--dest", "../Makefile2")

    cmd.Env = append(os.Environ(),
        "BLOBSTORE_READ_ACL=" + testingAccessToken,
        "BLOBSTORE_WRITE_ACL=" + testingAccessToken,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        assert.Failf(t, err.Error(), string(output))
    }

    assert.Nil(t, err)
    assert.Equal(t, "", string(output))

    receivedFile, err := os.Open("../Makefile2")
    assert.Nil(t, err)

    receivedBody, err := ioutil.ReadAll(bufio.NewReader(receivedFile))
    assert.Nil(t, err)

    expectedFile, err := os.Open("../Makefile")
    assert.Nil(t, err)

    expectedBody, err := ioutil.ReadAll(bufio.NewReader(expectedFile))
    assert.Nil(t, err)

    assert.Equal(t, expectedBody, receivedBody)
}

func TestCommandLineInterfaceDownloadFails(t *testing.T) {
    cmd := exec.Command("blob", "download", "--filename", "clientlib/testing/Makefile", "--dest", "../Makefile2")

    output, err := cmd.CombinedOutput()
    if err == nil {
        assert.Fail(t, "Expected a failure from download command")
    }

    expectedOutput := `Blobstore Download Failed (404): {"code":"NotFound","message":"File not found"}` + "\n" + *blobCliHelpString + "\n"
    assert.Equal(t, expectedOutput, string(output))
}
