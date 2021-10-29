package credential_provider

import (
	"net/http"
)

type ICredentialProvider interface {
	AuthorizeRequest(*http.Request) error
}
