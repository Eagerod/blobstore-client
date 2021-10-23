package blob

type BlobFileStat struct {
	Path      string
	Name      string
	MimeType  string
	SizeBytes int
	Exists    bool
}
