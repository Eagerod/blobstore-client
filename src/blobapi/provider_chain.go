package blobapi;

import (
    "net/http"
    "net/textproto"
    "os"
)

const (
    HttpRequestReadAclHeader = "X-BlobStore-Read-Acl"
    HttpRequestWriteAclHeader = "X-BlobStore-Write-Acl"

    DefaultBlobStoreReadAclEnvironmentVariable = "BLOBSTORE_READ_ACL"
    DefaultBlobStoreWriteAclEnvironmentVariable = "BLOBSTORE_WRITE_ACL"
)

type CredentialProviderChain struct {
    providers []ICredentialProvider
}

var defaultPc *CredentialProviderChain 

type ICredentialProvider interface {
    AuthorizeRequest(*http.Request) error
}

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

type DirectCredentialProvider struct {
    ReadAcl string
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

type EnvironmentCredentialProvider struct {
    ReadAclEnvironmentVariable string
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
