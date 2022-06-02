package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

import (
	"github.com/Eagerod/blobstore-client/pkg/blob"
	"github.com/Eagerod/blobstore-client/pkg/credential_provider"
)

// NOTE: This package expects a fresh binary to have just been installed on the
// machine. This should be updated to either use source, or use the binary in
// the project's build directory if possible, but that may be more difficult
// than needed.
const testingAccessToken string = "ad4c3f2d4fb81f4118f837464b961eebda026d8c52a7cc967047cc3c2a3f6a43"
const makefilePath string = "Makefile"
const blobBinPath string = "./build/blob"
const blobstoreBaseUrl string = "https://blob.aleemhaji.com"

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

func getTestFilePath() string {
	suffix := uuid.New().String()
	return path.Join("clientlib", "testing", suffix)
}

func getTestFileCliPath(base string) string {
	return path.Join("blob:", base)
}

func toURL(path string) *url.URL {
	url, err := url.Parse(path)
	if err != nil {
		panic(err)
	}

	return url
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
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteCliPath, "--type", "text/plain", "--force")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	contents, err := api.GetFileContents(toURL(remotePath))
	assert.Nil(t, err)
	defer api.DeleteFile(toURL(remotePath))

	assert.Equal(t, string(*makefileBytes), contents)
}

func TestCommandLineInterfaceUploadNoContentType(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteCliPath, "--force")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	contents, err := api.GetFileContents(toURL(remotePath))
	assert.Nil(t, err)
	defer api.DeleteFile(toURL(remotePath))

	assert.Equal(t, string(*makefileBytes), contents)
}

func TestCommandLineInterfaceUploadAlreadyExists(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	err := api.UploadFile(toURL(remotePath), makefilePath, "text/plain")
	assert.Nil(t, err)
	defer api.DeleteFile(toURL(remotePath))

	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteCliPath, "--type", "text/plain")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := "Error: Destination file already exists on blobstore; use --force to overwrite\n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceUploadFails(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	err := api.UploadFile(toURL(remotePath), makefilePath, "text/plain")
	assert.Nil(t, err)
	defer api.DeleteFile(toURL(remotePath))

	cmd := exec.Command(blobBinPath, "cp", makefilePath, remoteCliPath, "--type", "text/plain", "--force")
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := "Error: Blobstore Upload Failed (403): \n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceDownload(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(toURL(remotePath), makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "cp", remoteCliPath, "../Makefile2")
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
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(toURL(remotePath), makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "cp", remoteCliPath)
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, append(*makefileBytes, []byte("\n")...), output)
}

func TestCommandLineInterfaceDownloadFileAlreadyExists(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	cmd := exec.Command(blobBinPath, "cp", remoteCliPath, makefilePath)
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := `Error: Destination file already exists on local machine; use --force to overwrite` + "\n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceDownloadFails(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	cmd := exec.Command(blobBinPath, "cp", remoteCliPath, "../Makefile2")
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from download command")
	}

	expectedOutput := `Error: Blobstore Download Failed (404): {"code":"NotFound","message":"File not found"}` + "\n" + blobCliHelpStrings["cp"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceAppend(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(toURL(remotePath), makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "append", remoteCliPath, "--string", "something extra")
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	contents, err := api.GetFileContents(toURL(remotePath))
	assert.Nil(t, err)

	assert.Equal(t, string(*makefileBytes)+"something extra", contents)
}

func TestCommandLineInterfaceAppendFails(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	cmd := exec.Command(blobBinPath, "append", remoteCliPath, "--string", "something extra")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from append command")
	}

	expectedOutput := `Error: Blobstore Download Failed (404): {"code":"NotFound","message":"File not found"}` + "\n" + blobCliHelpStrings["append"] + "\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestCommandLineInterfaceList(t *testing.T) {
	remotePath := getTestFilePath()

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(toURL(remotePath), makefilePath, "text/plain")

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
	remotePath := getTestFilePath()

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(toURL(remotePath), makefilePath, "text/plain")

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
		if str == remotePath {
			found = true
			break
		}
	}

	assert.True(t, found, fmt.Sprintf("Did not find %s in blobstorage.", remotePath))
}

func TestCommandLineInterfaceDelete(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	api.UploadFile(toURL(remotePath), makefilePath, "text/plain")

	cmd := exec.Command(blobBinPath, "rm", remoteCliPath)
	cmd.Env = makeEnv(testingAccessToken)

	output, err := cmd.CombinedOutput()
	if err != nil {
		assert.Failf(t, err.Error(), string(output))
	}

	assert.Nil(t, err)
	assert.Equal(t, "", string(output))

	stat, err := api.StatFile(toURL(remotePath))
	assert.Nil(t, err)

	assert.Equal(t, stat.Exists, false)
}

func TestCommandLineInterfaceDeleteFails(t *testing.T) {
	remotePath := getTestFilePath()
	remoteCliPath := getTestFileCliPath(remotePath)

	api := blob.NewBlobStoreClient(blobstoreBaseUrl, &credential_provider.DirectCredentialProvider{testingAccessToken, testingAccessToken})
	err := api.UploadFile(toURL(remotePath), makefilePath, "text/plain")
	assert.Nil(t, err)
	defer api.DeleteFile(toURL(remotePath))

	cmd := exec.Command(blobBinPath, "rm", remoteCliPath)
	cmd.Env = makeEnv("")

	output, err := cmd.CombinedOutput()
	if err == nil {
		assert.Fail(t, "Expected a failure from delete command")
	}

	expectedOutput := "Error: Blobstore Delete Failed (403): \n" + blobCliHelpStrings["rm"] + "\n"
	assert.Equal(t, expectedOutput, string(output))

	stat, err := api.StatFile(toURL(remotePath))
	assert.Nil(t, err)

	assert.Equal(t, stat.Exists, true)
}
