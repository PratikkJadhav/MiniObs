package storage

import (
	"encoding/binary"
	"io"
	"os"
)

type segment struct {
	fileID uint32
	file   *os.File
	offset int64
}

func openSegment(fileID uint32, path string) (*segment, error) {

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	return &segment{
		fileID: fileID,
		file:   file,
		offset: offset,
	}, nil
}

func (s *segment) Write(data []byte) (location, error) {

	buf := make([]byte, 4+len(data))

	startOffset := s.offset

	binary.BigEndian.PutUint32(buf[0:4], uint32(len(data)))

	copy(buf[4:], data)

	_, err := s.file.Write(buf)

	s.offset += int64(4 + len(data))

	l := location{
		fileID: s.fileID,
		offset: startOffset,
		size:   uint32(len(data)),
	}
	return l, err
}

func (s *segment) Read(loc location) ([]byte, error) {
	file, err := os.OpenFile(s.file.Name(), os.O_RDONLY, 0644)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	offset := loc.offset
	file.Seek(offset, 0)

	size := make([]byte, 4)
	_, err = io.ReadFull(file, size)
	x := binary.BigEndian.Uint32(size)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, x)
	_, err = io.ReadFull(file, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil

}
