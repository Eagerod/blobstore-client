package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

import (
	"gitea.internal.aleemhaji.com/aleem/blobapi/cmd/blobapi"
)

// NOTE: This package expects a fresh binary to have just been installed on the
// machine. This should be updated to either use source, or use the binary in
// the project's build directory if possible, but that may be more difficult
// than needed.
const testingAccessToken string = "ad4c3f2d4fb81f4118f837464b961eebda026d8c52a7cc967047cc3c2a3f6a43"
const makefilePath string = "Makefile"
const remoteMakefileCliPath string = "blob:/clientlib/testing/Makefile"
const remoteMakefileRelPath string = "clientlib/testing/Makefile"
const blobBinPath string = "./build/blob"
const blobstoreBaseUrl string = "https://blob.internal.aleemhaji.com"

var commands []string = make([]string, 0, 0)
var blobCliHelpStrings map[string]string = make(map[string]string, 0)
var makefileBytes *[]byte

func makeEnv(withToken string) []string {
	env := os.Environ()
	copyEnv := make([]string, 0, len(env))
	for i := range env {
		if strings.HasPrefix(env[i], "BLOBSTORE_READ_ACL=") || strings.HasPrefix(env[i], "BLOBSTORE_WRITE_ACL=") {
			continue
		}
		copyEnv = append(copyEnv, env[i])
	}

	copyEnv = append(copyEnv, "BLOBSTORE_READ_ACL="+withToken, "BLOBSTORE_WRITE_ACL="+withToken)

	return copyEnv
}

func TestMain(m *testing.M) {
	if _, err := exec.LookPath(blobBinPath); err != nil {
		panic("Failed to find executable to run system tests")
	}

	commands = append(commands, "", "cp", "append", "rm")

	for i := range commands {
		command := commands[i]
		cmd := exec.Command(blobBinPath, command, "-h")

		str, err := cmd.CombinedOutput()
		if err != nil {
			panic("Failed to get help text from blob binary")
		}

		// Each explicit help output adds the command's description to the output
		// so remove the first line.
		stringPieces := strings.SplitN(string(str), "\n\n", 2)
		blobCliHelpStrings[command] = stringPieces[1]
	}

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
	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteMakefileCliPath, "--type", "text/plain", "--force")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	contents, err := api.GetFileContents(remoteMakefileRelPath)
	assert.Nil(t, err)

	assert.Equal(t, string(*makefileBytes), contents)
}

func TestCommandLineInterfaceUploadNoContentType(t *testing.T) {
	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteMakefileCliPath, "--force")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	contents, err := api.GetFileContents(remoteMakefileRelPath)
	assert.Nil(t, err)

	assert.Equal(t, string(*makefileBytes), contents)
}

func TestCommandLineInterfaceUploadAlreadyExists(t *testing.T) {
	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteMakefileCliPath, "--type", "text/plain")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := "Error: Destination file already exists on blobstore; use --force to overwrite\n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceUploadFails(t *testing.T) {
	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteMakefileCliPath, "--type", "text/plain", "--force")
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := "Error: Blobstore Upload Failed (403): \n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceDownload(t *testing.T) {
	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(remoteMakefileRelPath, makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "cp", remoteMakefileCliPath, "../Makefile2")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}
	defer os.Remove("../Makefile2")

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	receivedFile, err := os.Open("../Makefile2")
	assert.Nil(t, err)

	receivedBody, err := ioutil.ReadAll(bufio.NewReader(receivedFile))
	assert.Nil(t, err)

	assert.Equal(t, *makefileBytes, receivedBody)
}

func TestCommandLineInterfaceDownloadToSdtout(t *testing.T) {
	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(remoteMakefileRelPath, makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "cp", remoteMakefileCliPath)
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, append(*makefileBytes, []byte("\n")...), output)
}

func TestCommandLineInterfaceDownloadFileAlreadyExists(t *testing.T) {
	cmd := exec.Command(blobBinPath, "cp", remoteMakefileCliPath, makefilePath)
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := `Error: Destination file already exists on local machine; use --force to overwrite` + "\n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceDownloadFails(t *testing.T) {
	cmd := exec.Command(blobBinPath, "cp", remoteMakefileCliPath, "../Makefile2")
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := `Error: Blobstore Download Failed (404): {"code":"NotFound","message":"File not found"}` + "\n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceAppend(t *testing.T) {
	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(remoteMakefileRelPath, makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "append", remoteMakefileCliPath, "--string", "something extra")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	contents, err := api.GetFileContents(remoteMakefileRelPath)
	assert.Nil(t, err)

	assert.Equal(t, string(*makefileBytes)+"something extra", contents)
}

func TestCommandLineInterfaceAppendFails(t *testing.T) {
	cmd := exec.Command(blobBinPath, "append", remoteMakefileCliPath, "--string", "something extra")
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from append command")
	}

	expectedOutput := `Error: Blobstore Download Failed (404): {"code":"NotFound","message":"File not found"}` + "\n" + blobCliHelpStrings["append"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceList(t *testing.T) {
	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(remoteMakefileRelPath, makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "ls", "blob:/clientlib")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)

	foundFiles := strings.Split(string(output), "\n")
	found := false
	for _, str := range foundFiles {
		if str == "clientlib/testing/" {
			found = true
			break
		}
	}

	assert.True(t, found, "Did not find clientlib/testing/ in blobstorage.")
}

func TestCommandLineInterfaceListRecursive(t *testing.T) {
	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(remoteMakefileRelPath, makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "ls", "blob:/clientlib", "-r")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)

	foundFiles := strings.Split(string(output), "\n")
	found := false
	for _, str := range foundFiles {
		if str == "clientlib/testing/makefile" {
			found = true
			break
		}
	}

	assert.True(t, found, "Did not find clientlib/testing/makefile in blobstorage.")
}

func TestCommandLineInterfaceDelete(t *testing.T) {
	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(remoteMakefileRelPath, makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "rm", remoteMakefileCliPath)
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	stat, err := api.StatFile(remoteMakefileRelPath)
	assert.Nil(t, err)

	assert.Equal(t, stat.Exists, false)
}

func TestCommandLineInterfaceDeleteFails(t *testing.T) {
	api := blobapi.NewBlobStoreApiClient(blobstoreBaseUrl, &blobapi.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(remoteMakefileRelPath, makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "rm", remoteMakefileCliPath)
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from delete command")
	}

	expectedOutput := "Error: Blobstore Delete Failed (403): \n" + blobCliHelpStrings["rm"] + "\n"
	assert.Equal(t, expectedOutput, string(output))

	stat, err := api.StatFile(remoteMakefileRelPath)
	assert.Nil(t, err)

	assert.Equal(t, stat.Exists, true)
}
