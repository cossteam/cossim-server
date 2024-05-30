package qr

import (
	"bytes"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// WriteCloserWrapper wraps a bytes.Buffer to implement io.WriteCloser
type WriteCloserWrapper struct {
	buffer *bytes.Buffer
}

// NewWriteCloserWrapper creates a new WriteCloserWrapper
func NewWriteCloserWrapper() *WriteCloserWrapper {
	return &WriteCloserWrapper{
		buffer: new(bytes.Buffer),
	}
}

// Write writes data to the buffer
func (w *WriteCloserWrapper) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

// Close does nothing in this case, satisfies io.WriteCloser interface
func (w *WriteCloserWrapper) Close() error {
	return nil
}

func GenQrcode(text string) (*bytes.Buffer, error) {
	qrc, err := qrcode.New(text)
	if err != nil {
		return nil, err
	}

	// Create a new WriteCloserWrapper
	wc := NewWriteCloserWrapper()

	// Use WriteCloserWrapper as Writer
	writer := standard.NewWithWriter(wc)

	// Save the QR code to the buffer
	if err = qrc.Save(writer); err != nil {
		return nil, err
	}
	return wc.buffer, nil
}
