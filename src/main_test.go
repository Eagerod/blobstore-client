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
const makefilePath string = "../Makefile"
const remoteMakefilePath string = "clientlib/testing/Makefile"

var blobCliHelpString *string
var makefileBytes *[]byte

func TestMain(m *testing.M) {
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

    file, err := os.Open(makefilePath)
    if err != nil {
        panic("Failed to find Makefile for upload tests.")
    }

    expectedBody, err := ioutil.ReadAll(bufio.NewReader(file))
    if err != nil {
        panic("Failed to read bytes out from Makefile")
    }

    makefileBytes = &expectedBody

    retCode := m.Run()

    os.Remove("../Makefile2")
    os.Exit(retCode)
}

func TestCommandLineInterfaceUpload(t *testing.T) {
    cmd := exec.Command("blob", "upload", "--filename", remoteMakefilePath, "--type", "text/plain", "--source", makefilePath)

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
    contents, err := api.GetFileContents(remoteMakefilePath)
    assert.Nil(t, err)

    assert.Equal(t, string(*makefileBytes), contents)
}

func TestCommandLineInterfaceUploadNoContentType(t *testing.T) {
    cmd := exec.Command("blob", "upload", "--filename", remoteMakefilePath, "--source", makefilePath)

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
    contents, err := api.GetFileContents(remoteMakefilePath)
    assert.Nil(t, err)

    assert.Equal(t, string(*makefileBytes), contents)
}

func TestCommandLineInterfaceUploadFails(t *testing.T) {
    cmd := exec.Command("blob", "upload", "--filename", remoteMakefilePath, "--type", "text/plain", "--source", makefilePath)

    output, err := cmd.CombinedOutput()
    if err == nil {
        assert.Fail(t, "Expected a failure from download command")
    }

    expectedOutput := "Blobstore Upload Failed (403): \n" + *blobCliHelpString + "\n"
    assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceDownload(t *testing.T) {
    api := blobapi.NewBlobStoreApiClient("https://aleem.haji.ca/blob", testingAccessToken, testingAccessToken)
    api.UploadFile(remoteMakefilePath, makefilePath, "text/plain")

    cmd := exec.Command("blob", "download", "--filename", remoteMakefilePath, "--dest", "../Makefile2")

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

    assert.Equal(t, *makefileBytes, receivedBody)
}

func TestCommandLineInterfaceDownloadToSdtout(t *testing.T) {
    api := blobapi.NewBlobStoreApiClient("https://aleem.haji.ca/blob", testingAccessToken, testingAccessToken)
    api.UploadFile(remoteMakefilePath, makefilePath, "text/plain")

    cmd := exec.Command("blob", "download", "--filename", remoteMakefilePath)

    cmd.Env = append(os.Environ(),
        "BLOBSTORE_READ_ACL=" + testingAccessToken,
        "BLOBSTORE_WRITE_ACL=" + testingAccessToken,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        assert.Failf(t, err.Error(), string(output))
    }

    assert.Nil(t, err)
    assert.Equal(t, append(*makefileBytes, []byte("\n")...), output)
}

func TestCommandLineInterfaceDownloadFails(t *testing.T) {
    cmd := exec.Command("blob", "download", "--filename", remoteMakefilePath, "--dest", "../Makefile2")

    output, err := cmd.CombinedOutput()
    if err == nil {
        assert.Fail(t, "Expected a failure from download command")
    }

    expectedOutput := `Blobstore Download Failed (404): {"code":"NotFound","message":"File not found"}` + "\n" + *blobCliHelpString + "\n"
    assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceAppend(t *testing.T) {
    api := blobapi.NewBlobStoreApiClient("https://aleem.haji.ca/blob", testingAccessToken, testingAccessToken)
    api.UploadFile(remoteMakefilePath, makefilePath, "text/plain")

    cmd := exec.Command("blob", "append", "--filename", remoteMakefilePath, "--string", "something extra")

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

    contents, err := api.GetFileContents(remoteMakefilePath)
    assert.Nil(t, err)

    assert.Equal(t, string(*makefileBytes) + "something extra", contents)
}

func TestCommandLineInterfaceAppendFails(t *testing.T) {
    cmd := exec.Command("blob", "append", "--filename", remoteMakefilePath, "--string", "something extra")

    output, err := cmd.CombinedOutput()
    if err == nil {
        assert.Fail(t, "Expected a failure from append command")
    }

    expectedOutput := `Blobstore Download Failed (404): {"code":"NotFound","message":"File not found"}` + "\n" + *blobCliHelpString + "\n"
    assert.Equal(t, expectedOutput, string(output))
}
