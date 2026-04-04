package test

import (
	"bytes"
)

type MockFile struct {
	data   *bytes.Reader
	buffer *bytes.Buffer
}

func NewMockFile(data []byte) *MockFile {
	return &MockFile{
		data:   bytes.NewReader(data),
		buffer: bytes.NewBuffer(data),
	}
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	return m.data.Read(p)
}

func (m *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	return m.data.ReadAt(p, off)
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	return m.data.Seek(offset, whence)
}

func (m *MockFile) Close() error {
	return nil
}