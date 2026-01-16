package encoder

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

// Compress and encode
func CompressAndEncode(yaml []byte) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(yaml); err != nil {
		return "", err
	}
	gz.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// Decode and decompress
func DecodeAndDecompress(b64 string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return io.ReadAll(gz)
}
