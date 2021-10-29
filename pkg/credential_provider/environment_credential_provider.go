package credential_provider

import (
	"net/http"
	"os"
)

type EnvironmentCredentialProvider struct {
	ReadAclEnvironmentVariable  string
	WriteAclEnvironmentVariable string
}

func (ecp *EnvironmentCredentialProvider) AuthorizeRequest(request *http.Request) error {
	if !HasReadAclHeader(request) {
		if acl, ok := os.LookupEnv(ecp.ReadAclEnvironmentVariable); ok {
			request.Header.Add(HttpRequestReadAclHeader, acl)
		}
	}
	if !HasWriteAclHeader(request) {
		if acl, ok := os.LookupEnv(ecp.WriteAclEnvironmentVariable); ok {
			request.Header.Add(HttpRequestWriteAclHeader, acl)
		}
	}
	return nil
}
