package blob

import (
	"bufio"
	// "encoding/json"
	// "errors"
	// "fmt"
	"io"
	// "io/ioutil"
	// "net/http"
	// "net/url"
	// "os"
	// "path/filepath"
	// "strconv"
	// "strings"
	// "time"
)

// import (
// 	"gitea.internal.aleemhaji.com/aleem/blobapi/pkg/credential_provider"
// )

type BlobFileStat struct {
	Path      string
	Name      string
	MimeType  string
	SizeBytes int
	Exists    bool
}

type IBlobStoreApiClient interface {
	UploadStream(path string, stream *bufio.Reader, contentType string) error

	GetStat(path string) (*BlobFileStat, error)
	GetStream(path string) (*io.Reader, error)

	ListPrefix(prefix string, recursive bool) ([]string, error)

	DeleteFile(path string) error
}

type BlobStoreApiClient struct {

}

