package parser

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func buildHFSCatalogRecord(parentID uint32, name string, data []byte) []byte {
	units := utf16.Encode([]rune(name))
	key := new(bytes.Buffer)
	_ = binary.Write(key, binary.BigEndian, uint32(parentID))
	_ = binary.Write(key, binary.BigEndian, uint16(len(units)))
	for _, u := range units {
		_ = binary.Write(key, binary.BigEndian, u)
	}

	rec := new(bytes.Buffer)
	_ = binary.Write(rec, binary.BigEndian, uint16(key.Len()))
	rec.Write(key.Bytes())
	rec.Write(data)
	return rec.Bytes()
}

func buildHFSFolderRecord(folderID uint32) []byte {
	data := make([]byte, 12)
	binary.BigEndian.PutUint16(data[0:2], kHFSPlusFolderRecord)
	binary.BigEndian.PutUint32(data[8:12], folderID)
	return data
}

func buildHFSFileRecord(fileID uint32, dataForkStartBlock, dataForkBlockCount uint32, logicalSize uint64) []byte {
	data := make([]byte, 168)
	binary.BigEndian.PutUint16(data[0:2], kHFSPlusFileRecord)
	binary.BigEndian.PutUint32(data[8:12], fileID)

	binary.BigEndian.PutUint64(data[88:96], logicalSize)
	binary.BigEndian.PutUint32(data[104:108], dataForkStartBlock)
	binary.BigEndian.PutUint32(data[108:112], dataForkBlockCount)
	return data
}

func buildHFSLeafNode(nodeSize int, records [][]byte) []byte {
	node := make([]byte, nodeSize)

	node[8] = 0xFF
	node[9] = 1
	binary.BigEndian.PutUint16(node[10:12], uint16(len(records)))

	offsets := make([]uint16, 0, len(records)+1)
	pos := 14
	for _, rec := range records {
		offsets = append(offsets, uint16(pos))
		copy(node[pos:], rec)
		pos += len(rec)
	}
	offsets = append(offsets, uint16(pos))

	tableStart := nodeSize - 2*len(offsets)
	for i, off := range offsets {
		binary.BigEndian.PutUint16(node[tableStart+2*i:], off)
	}
	return node
}

func buildHFSHeaderNode(nodeSize, totalNodes, firstLeafNode int) []byte {
	node := make([]byte, nodeSize)
	binary.BigEndian.PutUint16(node[10:12], 3)

	binary.BigEndian.PutUint32(node[16:20], 1)
	binary.BigEndian.PutUint32(node[20:24], 1)
	binary.BigEndian.PutUint32(node[24:28], uint32(firstLeafNode))
	binary.BigEndian.PutUint32(node[28:32], uint32(firstLeafNode))
	binary.BigEndian.PutUint16(node[32:34], uint16(nodeSize))
	binary.BigEndian.PutUint32(node[36:40], uint32(totalNodes))
	return node
}

func buildSyntheticHFSPlusVolume(t *testing.T, plist []byte) []byte {
	t.Helper()
	const blockSize = 4096

	catalogNodes := buildHFSHeaderNode(blockSize, 2, 1)
	leaf := buildHFSLeafNode(blockSize, [][]byte{
		buildHFSCatalogRecord(hfsRootFolderID, "Test.app", buildHFSFolderRecord(10)),
		buildHFSCatalogRecord(10, "Contents", buildHFSFolderRecord(11)),
		buildHFSCatalogRecord(11, "Info.plist", buildHFSFileRecord(12, 3, 1, uint64(len(plist)))),
	})
	catalogNodes = append(catalogNodes, leaf...)

	img := make([]byte, blockSize*4)
	copy(img[3*blockSize:], plist)

	header := make([]byte, 512)
	binary.BigEndian.PutUint16(header[0:2], hfsPlusSignature)
	binary.BigEndian.PutUint32(header[40:44], blockSize)

	binary.BigEndian.PutUint64(header[272:280], uint64(len(catalogNodes)))
	binary.BigEndian.PutUint32(header[272+16:272+20], 1)
	binary.BigEndian.PutUint32(header[272+20:272+24], 2)
	copy(img[1024:1024+512], header)

	copy(img[blockSize:], catalogNodes)

	return img
}

func TestReadHFSPlusInfoPlist_Synthetic(t *testing.T) {
	plist := []byte(`<?xml version="1.0"?><plist><dict><key>CFBundleIdentifier</key><string>com.example.test</string></dict></plist>`)
	img := buildSyntheticHFSPlusVolume(t, plist)

	data, ok := readHFSPlusInfoPlist(bytes.NewReader(img))
	if !ok {
		t.Fatal("readHFSPlusInfoPlist() ok = false, want true")
	}
	if !bytes.Equal(data, plist) {
		t.Errorf("readHFSPlusInfoPlist() = %q, want %q", data, plist)
	}
}

func TestReadHFSPlusInfoPlist_NotHFSPlus(t *testing.T) {
	garbage := make([]byte, 4096)
	if _, ok := readHFSPlusInfoPlist(bytes.NewReader(garbage)); ok {
		t.Error("readHFSPlusInfoPlist() ok = true for non-HFS+ data, want false")
	}
}
