package blobapi;

import (
    "net/http"
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
    _, ok := request.Header[HttpRequestReadAclHeader]
    return ok
}

func HasWriteAclHeader(request *http.Request) bool {
    _, ok := request.Header[HttpRequestWriteAclHeader]
    return ok
}

func DefaultCredentialProviderChain() *CredentialProviderChain {
    if defaultPc != nil {
        return defaultPc
    }

    defaultPc = new(CredentialProviderChain)

    defaultPc.providers = append(defaultPc.providers,
        &EvironmentCredentialProvider{
            DefaultBlobStoreReadAclEnvironmentVariable,
            DefaultBlobStoreWriteAclEnvironmentVariable,
        },
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

type EvironmentCredentialProvider struct {
    ReadAclEnvironmentVariable string
    WriteAclEnvironmentVariable string
}

func (ecp *EvironmentCredentialProvider) AuthorizeRequest(request *http.Request) error {
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
