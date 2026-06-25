package parser

import (
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"unicode/utf16"
)

const (
	hfsPlusSignature = 0x482B
	hfsXSignature    = 0x4858
	hfsRootFolderID  = 2

	kHFSPlusFolderRecord       = 1
	kHFSPlusFileRecord         = 2
	kHFSPlusFolderThreadRecord = 3
	kHFSPlusFileThreadRecord   = 4
)

var errHFSPlusUnsupportedShape = errors.New("unsupported or fragmented HFS+ structure")

type hfsExtent struct {
	startBlock uint32
	blockCount uint32
}

type hfsFork struct {
	logicalSize uint64
	extents     [8]hfsExtent
}

func readHFSPlusInfoPlist(r io.ReaderAt) (data []byte, ok bool) {
	defer func() {
		if recover() != nil {
			data, ok = nil, false
		}
	}()

	header := make([]byte, 512)
	if _, err := r.ReadAt(header, 1024); err != nil {
		return nil, false
	}

	sig := binary.BigEndian.Uint16(header[0:2])
	if sig != hfsPlusSignature && sig != hfsXSignature {
		return nil, false
	}

	blockSize := binary.BigEndian.Uint32(header[40:44])
	if blockSize == 0 {
		return nil, false
	}
	catalogFork := parseForkData(header[272:352])

	vol := &hfsVolume{r: r, blockSize: uint64(blockSize)}

	catalog, err := vol.readCatalog(catalogFork)
	if err != nil {
		return nil, false
	}

	appPaths := catalog.findAppBundles()
	for _, appID := range appPaths {
		contentsID, ok := catalog.childByName(appID, "Contents")
		if !ok {
			continue
		}
		plistEntry, ok := catalog.fileByName(contentsID, "Info.plist")
		if !ok {
			continue
		}
		data, err := vol.readFork(plistEntry.dataFork)
		if err != nil || len(data) == 0 {
			continue
		}
		return data, true
	}

	return nil, false
}

type hfsVolume struct {
	r         io.ReaderAt
	blockSize uint64
}

func (v *hfsVolume) readFork(fork hfsFork) ([]byte, error) {
	buf := make([]byte, 0, fork.logicalSize)
	var covered uint64

	for _, ext := range fork.extents {
		if ext.blockCount == 0 {
			continue
		}
		chunk := make([]byte, uint64(ext.blockCount)*v.blockSize)
		if _, err := v.r.ReadAt(chunk, int64(uint64(ext.startBlock)*v.blockSize)); err != nil {
			return nil, err
		}
		buf = append(buf, chunk...)
		covered += uint64(ext.blockCount) * v.blockSize
	}

	if covered < fork.logicalSize {
		return nil, errHFSPlusUnsupportedShape
	}
	if uint64(len(buf)) > fork.logicalSize {
		buf = buf[:fork.logicalSize]
	}
	return buf, nil
}

func parseForkData(b []byte) hfsFork {
	var f hfsFork
	f.logicalSize = binary.BigEndian.Uint64(b[0:8])
	for i := range f.extents {
		off := 16 + i*8
		f.extents[i] = hfsExtent{
			startBlock: binary.BigEndian.Uint32(b[off : off+4]),
			blockCount: binary.BigEndian.Uint32(b[off+4 : off+8]),
		}
	}
	return f
}

type catalogEntry struct {
	cnid     uint32
	parentID uint32
	name     string
	isFolder bool
	dataFork hfsFork
}

type hfsCatalog struct {
	childrenByParent map[uint32][]catalogEntry
}

func (v *hfsVolume) readCatalog(fork hfsFork) (*hfsCatalog, error) {
	raw, err := v.readFork(fork)
	if err != nil {
		return nil, err
	}
	if len(raw) < 256 {
		return nil, errHFSPlusUnsupportedShape
	}

	numRecords := binary.BigEndian.Uint16(raw[10:12])
	if numRecords == 0 {
		return nil, errHFSPlusUnsupportedShape
	}
	nodeSize := binary.BigEndian.Uint16(raw[32:34])
	firstLeafNode := binary.BigEndian.Uint32(raw[24:28])
	totalNodes := binary.BigEndian.Uint32(raw[36:40])
	if nodeSize == 0 || uint64(totalNodes)*uint64(nodeSize) > uint64(len(raw))+uint64(nodeSize) {
		return nil, errHFSPlusUnsupportedShape
	}

	cat := &hfsCatalog{childrenByParent: make(map[uint32][]catalogEntry)}

	nodeIdx := firstLeafNode
	visited := make(map[uint32]bool)
	for nodeIdx != 0 && !visited[nodeIdx] {
		visited[nodeIdx] = true
		start := uint64(nodeIdx) * uint64(nodeSize)
		if start+uint64(nodeSize) > uint64(len(raw)) {
			break
		}
		node := raw[start : start+uint64(nodeSize)]
		next := decodeLeafNode(node, cat)
		nodeIdx = next
	}

	return cat, nil
}

func decodeLeafNode(node []byte, cat *hfsCatalog) uint32 {
	kind := int8(node[8])
	numRecords := binary.BigEndian.Uint16(node[10:12])
	fLink := binary.BigEndian.Uint32(node[0:4])

	if kind != -1 {
		return fLink
	}

	offsetTableStart := len(node) - 2*(int(numRecords)+1)
	if offsetTableStart < 14 {
		return fLink
	}

	for i := 0; i < int(numRecords); i++ {
		recOff := int(binary.BigEndian.Uint16(node[offsetTableStart+2*i : offsetTableStart+2*i+2]))
		if recOff < 14 || recOff >= len(node) {
			continue
		}

		keyLen := int(binary.BigEndian.Uint16(node[recOff : recOff+2]))
		keyStart := recOff + 2
		if keyStart+keyLen > len(node) || keyLen < 6 {
			continue
		}
		parentID := binary.BigEndian.Uint32(node[keyStart : keyStart+4])
		nameLen := int(binary.BigEndian.Uint16(node[keyStart+4 : keyStart+6]))
		nameStart := keyStart + 6
		if nameStart+nameLen*2 > len(node) {
			continue
		}
		name := decodeHFSUniStr(node[nameStart : nameStart+nameLen*2])

		dataStart := keyStart + keyLen
		if keyLen%2 == 1 {
			dataStart++
		}
		if dataStart+2 > len(node) {
			continue
		}
		recordType := int16(binary.BigEndian.Uint16(node[dataStart : dataStart+2]))

		switch recordType {
		case kHFSPlusFolderRecord:
			if dataStart+12 > len(node) {
				continue
			}
			folderID := binary.BigEndian.Uint32(node[dataStart+8 : dataStart+12])
			cat.add(catalogEntry{cnid: folderID, parentID: parentID, name: name, isFolder: true})
		case kHFSPlusFileRecord:
			if dataStart+168 > len(node) {
				continue
			}
			fileID := binary.BigEndian.Uint32(node[dataStart+8 : dataStart+12])
			dataFork := parseForkData(node[dataStart+88 : dataStart+168])
			cat.add(catalogEntry{cnid: fileID, parentID: parentID, name: name, isFolder: false, dataFork: dataFork})
		case kHFSPlusFolderThreadRecord, kHFSPlusFileThreadRecord:
		}
	}

	return fLink
}

func (c *hfsCatalog) add(e catalogEntry) {
	c.childrenByParent[e.parentID] = append(c.childrenByParent[e.parentID], e)
}

func (c *hfsCatalog) findAppBundles() []uint32 {
	var apps []uint32
	queue := []uint32{hfsRootFolderID}
	visited := map[uint32]bool{}

	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		if visited[parent] {
			continue
		}
		visited[parent] = true

		for _, child := range c.childrenByParent[parent] {
			if !child.isFolder {
				continue
			}
			if strings.HasSuffix(strings.ToLower(child.name), ".app") {
				apps = append(apps, child.cnid)
			}
			queue = append(queue, child.cnid)
		}
	}
	return apps
}

func (c *hfsCatalog) childByName(parent uint32, name string) (uint32, bool) {
	for _, child := range c.childrenByParent[parent] {
		if child.isFolder && strings.EqualFold(child.name, name) {
			return child.cnid, true
		}
	}
	return 0, false
}

func (c *hfsCatalog) fileByName(parent uint32, name string) (catalogEntry, bool) {
	for _, child := range c.childrenByParent[parent] {
		if !child.isFolder && strings.EqualFold(child.name, name) {
			return child, true
		}
	}
	return catalogEntry{}, false
}

func decodeHFSUniStr(b []byte) string {
	units := make([]uint16, len(b)/2)
	for i := range units {
		units[i] = binary.BigEndian.Uint16(b[i*2 : i*2+2])
	}
	return string(utf16.Decode(units))
}
