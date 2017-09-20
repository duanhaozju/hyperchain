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

func (sectionReader *SectionReader) ReadAt(sid int64) (int, []byte, error) {
	sectionReader.latestShard = sid - 1
	return sectionReader.ReadNext()
}

func (sectionReader *SectionReader) Close() {
	if sectionReader.fd != nil {
		sectionReader.fd.Close()
	}
}

func (sectionReader *SectionReader) check(sid int64) bool {
	if sid == 0 || sid > sectionReader.shardNum {
		return false
	}
	return true
}
