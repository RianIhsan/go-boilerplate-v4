package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/blacktop/go-apfs/pkg/disk/dmg"
)

type fixedPartitionReader struct {
	part    *dmg.Partition
	rawFile *io.SectionReader

	baseOnce sync.Once
	base     int64
}

func (f *fixedPartitionReader) chunkBase() int64 {
	f.baseOnce.Do(func() {
		base := int64(-1)
		for _, chk := range f.part.Chunks {
			if chk.DiskLength == 0 {
				continue
			}
			if base == -1 || int64(chk.DiskOffset) < base {
				base = int64(chk.DiskOffset)
			}
		}
		if base == -1 {
			base = 0
		}
		f.base = base
	})
	return f.base
}

func (f *fixedPartitionReader) ReadAt(p []byte, off int64) (int, error) {
	remaining := int64(len(p))
	curOff := off + f.chunkBase()
	written := 0

	for remaining > 0 {
		found := false
		for _, chk := range f.part.Chunks {
			chkStart := int64(chk.DiskOffset)
			chkEnd := chkStart + int64(chk.DiskLength)
			if curOff < chkStart || curOff >= chkEnd {
				continue
			}
			found = true

			var buf bytes.Buffer
			if _, err := chk.DecompressChunk(f.rawFile, make([]byte, chk.CompressedLength), &buf); err != nil {
				return written, err
			}
			data := buf.Bytes()
			diff := curOff - chkStart
			avail := int64(len(data)) - diff
			if avail <= 0 {
				return written, fmt.Errorf("decompressed chunk shorter than expected")
			}

			n := min(avail, remaining)
			copy(p[written:], data[diff:diff+n])
			written += int(n)
			curOff += n
			remaining -= n
			break
		}
		if !found {
			return written, io.ErrUnexpectedEOF
		}
	}

	return written, nil
}

func (f *fixedPartitionReader) Close() error { return nil }

func (f *fixedPartitionReader) ReadFile(w *bufio.Writer, off, length int64) error {
	buf := make([]byte, length)
	n, err := f.ReadAt(buf, off)
	if err != nil && err != io.EOF {
		return err
	}
	_, err = w.Write(buf[:n])
	return err
}

func (f *fixedPartitionReader) GetSize() uint64 {
	var max uint64
	for _, chk := range f.part.Chunks {
		if end := chk.DiskOffset + chk.DiskLength; end > max {
			max = end
		}
	}
	return max
}
