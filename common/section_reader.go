//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"errors"
	"io"
	"os"
)

type SectionReader struct {
	filePath    string
	shardLen    int64
	latestShard int64
	shardNum    int64
	fd          *os.File
}

var (
	InvalidShardIdErr = errors.New("invalid shard id")
	EmptyFileErr      = errors.New("empty file")
	NotFileErr        = errors.New("is not file")
)

// NewSectionReader returns a SectionReader, contains filePath and an open file.
func NewSectionReader(filePath string, shardLen int64) (error, *SectionReader) {
	fd, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return err, nil
	}
	fstat, err := fd.Stat()
	if err != nil {
		return err, nil
	}

	if fstat.IsDir() {
		return NotFileErr, nil
	}

	size := fstat.Size()
	if size == 0 {
		return EmptyFileErr, nil
	}

	shardNum := size / shardLen
	if (size % shardLen) > 0 {
		shardNum += 1
	}

	return nil, &SectionReader{
		filePath: filePath,
		shardLen: shardLen,
		fd:       fd,
		shardNum: shardNum,
	}
}

// ReadNext reads the next shard of the latestShard
func (sectionReader *SectionReader) ReadNext() (n int, buf []byte, err error) {
	defer func() {
		if err == nil {
			sectionReader.latestShard += 1
		}
	}()

	if !sectionReader.check(sectionReader.latestShard + 1) {
		n = 0
		buf = nil
		err = InvalidShardIdErr
		return
	}

	buf = make([]byte, sectionReader.shardLen)
	offset := sectionReader.latestShard * sectionReader.shardLen
	_reader := io.NewSectionReader(sectionReader.fd, offset, sectionReader.shardLen)
	n, err = _reader.Read(buf)
	return
}

// ReadAt reads the sidth shard
func (sectionReader *SectionReader) ReadAt(sid int64) (int, []byte, error) {
	sectionReader.latestShard = sid - 1
	return sectionReader.ReadNext()
}

// Close closes the open file
func (sectionReader *SectionReader) Close() {
	if sectionReader.fd != nil {
		sectionReader.fd.Close()
	}
}

// check checks that if sid is legal or not
func (sectionReader *SectionReader) check(sid int64) bool {
	if sid == 0 || sid > sectionReader.shardNum {
		return false
	}
	return true
}
