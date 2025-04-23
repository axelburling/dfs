package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"

	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/node/storage/crypto"
)

type Compression struct {
	crypto *crypto.Crypto
	log    *log.Logger
}

type Compress struct {
	crypto *crypto.Crypto
	file   *os.File
	log    *log.Logger
}

func (c *Compress) Write(p []byte) (int, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)

	size, err := w.Write(p)

	if err != nil {
		return 0, err
	}

	err = w.Close()

	if err != nil {
		return 0, err
	}
	
	reader, err := c.crypto.EncryptReader(&b)

	if err != nil {
		return 0, err
	}

	_, err = io.Copy(c.file, reader)

	return size, err
}

func (c *Compression) Decompress(file *os.File) (io.ReadCloser, error) {
	dr, err := c.crypto.Decrypt(file)

	if err != nil {
		return nil, err
	}
	r, err := gzip.NewReader(dr)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Compression) NewCompress(file *os.File) *Compress {
	return &Compress{
		crypto: c.crypto,
		file: file,
		log: c.log,
	}
}

func NewCompression(c *crypto.Crypto, log *log.Logger) *Compression {
	return &Compression{
		crypto: c,
		log: log,
	}
}