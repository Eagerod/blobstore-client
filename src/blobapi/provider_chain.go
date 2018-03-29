package blobapi;

import (
    "http"
)

type CredentialProviderChain struct {
    providers []*ICredentialProvider
}

type ICredentialProvider interface {
    AuthorizeRequest(*http.Request) error
}

var defaultPc *ProviderChain 

func DefaultCredentialProviderChain() *CredentialProviderChain {
    if defaultPc {
        return defaultPc
    }

    defaultPc = new(ProviderChain)
    // defaultPc.providers = append(defaultPc.providers)
    return defaultPc
}

func (*cpc CredentialProviderChain) AuthorizeRequest(request *http.Request) error {
    for _, provider := range cpc.providers {
        if err := provider.AuthorizeRequest(request); err == nil {
            return nil
        }
    }

    return errors.New("Failed to find any mechanism to authenticate requests to Blobstore")
}

type EvironmentCredentialProvider struct {
    readEnvVar string
    writeEnvVar string
}

func (ecp *EvironmentCredentialProvider) AuthorizeRequest(request *http.Request) error {
    
    request.Header.Add("X-BlobStore-Read-Acl", b.DefaultReadAcl)
    request.Header.Add("X-BlobStore-Write-Acl", b.DefaultWriteAcl)
}