package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"sort"
	"strings"

	apfspkg "github.com/blacktop/go-apfs"
	"github.com/blacktop/go-apfs/pkg/disk/dmg"
	"github.com/blacktop/go-apfs/types"
	"howett.net/plist"
)

var errUnexpectedAPFSShape = errors.New("unexpected apfs object map shape")

func ParseDMG(path string) (map[string]any, error) {
	d, err := dmg.Open(path, &dmg.Config{})
	if err != nil {
		if err == dmg.ErrEncrypted {
			return map[string]any{}, nil
		}
		return nil, err
	}
	defer d.Close()

	raw, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer raw.Close()
	fi, err := raw.Stat()
	if err != nil {
		return nil, err
	}
	rawSR := io.NewSectionReader(raw, 0, fi.Size())

	for i := range d.Partitions {
		dev := &fixedPartitionReader{part: &d.Partitions[i], rawFile: rawSR}

		data, ok := readAPFSInfoPlist(dev)
		if !ok {
			data, ok = readHFSPlusInfoPlist(dev)
		}
		if !ok {
			continue
		}

		var info map[string]any
		if _, err := plist.Unmarshal(data, &info); err != nil {
			continue
		}

		if len(info) > 0 {
			return info, nil
		}
	}

	return map[string]any{}, nil
}

func readAPFSInfoPlist(dev *fixedPartitionReader) (data []byte, ok bool) {
	defer func() {
		if recover() != nil {
			data, ok = nil, false
		}
	}()

	a, err := apfspkg.NewAPFS(dev)
	if err != nil {
		return nil, false
	}
	defer a.Close()

	omapTree, err := getOMapTree(a)
	if err != nil {
		return nil, false
	}

	sr := io.NewSectionReader(dev, 0, 1<<62)

	getChildRecords := func(oid uint64) (types.FSRecords, error) {
		return omapTree.GetFSRecordsForOid(sr, a.FSRootBtree, types.OidT(oid), types.XidT(^uint64(0)))
	}

	const maxDepth = 6
	queue := []uint64{types.FSROOT_OID}
	var appDirID uint64
	found := false

	for depth := 0; depth < maxDepth && len(queue) > 0 && !found; depth++ {
		var next []uint64
		for _, dirOid := range queue {
			records, err := getChildRecords(dirOid)
			if err != nil {
				continue
			}
			for _, rec := range records {
				if rec.Hdr.GetType() != types.APFS_TYPE_DIR_REC {
					continue
				}
				key, ok := rec.Key.(types.JDrecHashedKeyT)
				val, ok2 := rec.Val.(types.JDrecVal)
				if !ok || !ok2 {
					continue
				}
				if strings.HasSuffix(strings.ToLower(key.Name), ".app") {
					appDirID = val.FileID
					found = true
					break
				}
				next = append(next, val.FileID)
			}
			if found {
				break
			}
		}
		queue = next
	}
	if !found {
		return nil, false
	}

	appRecords, err := getChildRecords(appDirID)
	if err != nil {
		return nil, false
	}

	contentsDirID, found := findDirEntry(appRecords, func(name string) bool { return name == "Contents" })
	if !found {
		return nil, false
	}

	contentsRecords, err := getChildRecords(contentsDirID)
	if err != nil {
		return nil, false
	}

	plistFileID, found := findDirEntry(contentsRecords, func(name string) bool { return name == "Info.plist" })
	if !found {
		return nil, false
	}

	fileRecords, err := getChildRecords(plistFileID)
	if err != nil {
		return nil, false
	}

	data, err = readFileRecords(dev, fileRecords)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}

func getOMapTree(a *apfspkg.APFS) (types.BTreeNodePhys, error) {
	omap, ok := a.Volume.OMap.Body.(types.OMap)
	if !ok {
		return types.BTreeNodePhys{}, errUnexpectedAPFSShape
	}
	tree, ok := omap.Tree.Body.(types.BTreeNodePhys)
	if !ok {
		return types.BTreeNodePhys{}, errUnexpectedAPFSShape
	}
	return tree, nil
}

func findDirEntry(records types.FSRecords, match func(name string) bool) (uint64, bool) {
	for _, rec := range records {
		if rec.Hdr.GetType() != types.APFS_TYPE_DIR_REC {
			continue
		}
		key, ok := rec.Key.(types.JDrecHashedKeyT)
		if !ok || !match(key.Name) {
			continue
		}
		val, ok := rec.Val.(types.JDrecVal)
		if !ok {
			continue
		}
		return val.FileID, true
	}
	return 0, false
}

func readFileRecords(r io.ReaderAt, records types.FSRecords) ([]byte, error) {
	var fexts []types.FileExtent
	var totalBytesWritten uint64
	var decmpfsHdr *types.DecmpfsDiskHeader
	compressed := false

	for _, rec := range records {
		switch rec.Hdr.GetType() {
		case types.APFS_TYPE_INODE:
			inode, ok := rec.Val.(types.JInodeVal)
			if !ok {
				continue
			}
			if inode.InternalFlags&types.INODE_HAS_UNCOMPRESSED_SIZE != 0 {
				compressed = true
			}
			for _, xf := range inode.Xfields {
				if xf.XType == types.INO_EXT_TYPE_DSTREAM {
					if ds, ok := xf.Field.(types.JDstreamT); ok {
						totalBytesWritten = ds.TotalBytesWritten
					}
				}
			}
		case types.APFS_TYPE_FILE_EXTENT:
			key, ok1 := rec.Key.(types.JFileExtentKeyT)
			val, ok2 := rec.Val.(types.JFileExtentValT)
			if ok1 && ok2 {
				fexts = append(fexts, types.FileExtent{
					Address: key.LogicalAddr,
					Block:   val.PhysBlockNum,
					Length:  val.Length(),
				})
			}
		case types.APFS_TYPE_XATTR:
			xkey, ok := rec.Key.(types.JXattrKeyT)
			if !ok {
				continue
			}
			if xkey.Name == types.XATTR_DECMPFS_EA_NAME {
				if hdr, err := types.GetDecmpfsHeader(rec); err == nil {
					decmpfsHdr = hdr
				}
			}
		}
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	if compressed && decmpfsHdr != nil {
		if err := decmpfsHdr.DecompressFile(r, w, fexts, true); err != nil {
			return nil, err
		}
	} else {
		sort.Slice(fexts, func(i, j int) bool { return fexts[i].Address < fexts[j].Address })
		for _, fext := range fexts {
			sr := io.NewSectionReader(r, int64(fext.Block*types.BLOCK_SIZE), int64(fext.Length))
			if _, err := io.CopyN(w, sr, int64(fext.Length)); err != nil && err != io.EOF {
				return nil, err
			}
		}
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	out := buf.Bytes()
	if totalBytesWritten > 0 && uint64(len(out)) > totalBytesWritten {
		out = out[:totalBytesWritten]
	}
	return out, nil
}
