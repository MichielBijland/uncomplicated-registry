package storage

import (
	"context"
	"io"

	"github.com/pkg/errors"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Downloader struct {
	payload []byte
	error   bool
}

// Not 100% sure if that works correctly for large byte arrays
// Implements the s3DownloaderAPI interface
func (m *mockS3Downloader) Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(api *s3manager.Downloader)) (n int64, err error) {
	if m.error {
		return 0, errors.New("mocked error")
	}

	var off int64 = 0
	for {
		written, err := w.WriteAt(m.payload, off)
		if err != nil {
			return 0, err
		}
		off += int64(written)
		if off == int64(len(m.payload)) {
			break
		}
	}
	return 0, nil
}
