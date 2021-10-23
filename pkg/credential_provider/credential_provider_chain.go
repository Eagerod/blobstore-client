package credential_provider

import (
	"net/http"
	"net/textproto"
)

const (
	HttpRequestReadAclHeader  = "X-BlobStore-Read-Acl"
	HttpRequestWriteAclHeader = "X-BlobStore-Write-Acl"

	DefaultBlobStoreReadAclEnvironmentVariable  = "BLOBSTORE_READ_ACL"
	DefaultBlobStoreWriteAclEnvironmentVariable = "BLOBSTORE_WRITE_ACL"
)

type CredentialProviderChain struct {
	providers []ICredentialProvider
}

var defaultPc *CredentialProviderChain

func HasReadAclHeader(request *http.Request) bool {
	_, ok := request.Header[textproto.CanonicalMIMEHeaderKey(HttpRequestReadAclHeader)]
	return ok
}

func HasWriteAclHeader(request *http.Request) bool {
	_, ok := request.Header[textproto.CanonicalMIMEHeaderKey(HttpRequestWriteAclHeader)]
	return ok
}

func DefaultCredentialProviderChain() *CredentialProviderChain {
	if defaultPc != nil {
		return defaultPc
	}

	defaultPc = new(CredentialProviderChain)

	defaultPc.providers = append(defaultPc.providers,
		&EnvironmentCredentialProvider{
			DefaultBlobStoreReadAclEnvironmentVariable,
			DefaultBlobStoreWriteAclEnvironmentVariable,
		},
		&DirectCredentialProvider{"", ""},
	)
	return defaultPc
}

func (cpc *CredentialProviderChain) AuthorizeRequest(request *http.Request) error {
	for _, provider := range cpc.providers {
		if err := provider.AuthorizeRequest(request); err != nil {
			return err
		}
		if HasReadAclHeader(request) && HasWriteAclHeader(request) {
			return nil
		}
	}

	// Maybe no authorization is desired.
	return nil
}
