package credential_provider

import (
	"net/http"
)

type DirectCredentialProvider struct {
	ReadAcl  string
	WriteAcl string
}

func (dcp *DirectCredentialProvider) AuthorizeRequest(request *http.Request) error {
	if !HasReadAclHeader(request) {
		request.Header.Add(HttpRequestReadAclHeader, dcp.ReadAcl)
	}
	if !HasWriteAclHeader(request) {
		request.Header.Add(HttpRequestWriteAclHeader, dcp.WriteAcl)
	}
	return nil
}
