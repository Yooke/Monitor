package tools

import (
	"bytes"
	"github.com/SKatiyar/qr"
	"io"
)

func CreateQRCode(body string) (io.Reader, error) {
	code, err := qr.Encode(body, qr.L)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(code.PNG()), nil
}
