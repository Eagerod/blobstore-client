package credential_provider

import (
	"net/http"
	"os"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestDefaultProviderChainDefaultProviders(t *testing.T) {
	dcpc := DefaultCredentialProviderChain()

	assert.Equal(t, 2, len(dcpc.providers))
	assert.IsType(t, &EnvironmentCredentialProvider{}, dcpc.providers[0])
	assert.IsType(t, &DirectCredentialProvider{}, dcpc.providers[1])
}

func TestDefaultProviderChainDefaultProviderBehaviour(t *testing.T) {
	os.Setenv("BLOBSTORE_READ_ACL", "abc")
	os.Setenv("BLOBSTORE_WRITE_ACL", "bcd")

	defer func() {
		os.Unsetenv("BLOBSTORE_READ_ACL")
		os.Unsetenv("BLOBSTORE_WRITE_ACL")
	}()

	dcpc := DefaultCredentialProviderChain()

	request, err := http.NewRequest("GET", "https://example.org", nil)
	assert.Nil(t, err)

	err = dcpc.AuthorizeRequest(request)
	assert.Nil(t, err)
	assert.Equal(t, "abc", request.Header.Get("X-BlobStore-Read-Acl"))
	assert.Equal(t, "bcd", request.Header.Get("X-BlobStore-Write-Acl"))
}
