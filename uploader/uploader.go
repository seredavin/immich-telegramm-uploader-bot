package uploader

import (
	"io"
)

type Uploader interface {
	Upload(fileReader io.Reader, filename string, tags []string) (string, error)
}
