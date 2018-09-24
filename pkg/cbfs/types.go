package cbfs

import "encoding/binary"

type Props struct {
	Offset uint32
	Size   uint32
}

const (
	None uint32 = iota
	LZMA
	LZ4
)

var CBFSEndian = binary.BigEndian

// These are standard component types for well known
//   components (i.e - those that coreboot needs to consume.
//   Users are welcome to use any other value for their
//   components.
type CBFSFileType uint32

const (
	// FOV
	TypeDeleted2   CBFSFileType = 0xffffffff
	TypeDeleted                 = 0
	TypeStage                   = 0x10
	TypeSelf                    = 0x20
	TypeFIT                     = 0x21
	TypeOptionRom               = 0x30
	TypeBootSplash              = 0x40
	TypeRaw                     = 0x50
	TypeVSA                     = 0x51 // very, very obsolete Geode thing
	TypeMBI                     = 0x52
	TypeMicroCode               = 0x53
	TypeFSP                     = 0x60
	TypeMRC                     = 0x61
	TypeMMA                     = 0x62
	TypeEFI                     = 0x63
	TypeStruct                  = 0x70
	TypeCMOS                    = 0xaa
	TypeSPD                     = 0xab
	TypeMRCCache                = 0xac
	TypeCMOSLayout              = 0x1aa
)

const (
	HeaderMagic   = 0x4F524243
	HeaderV1      = 0x31313131
	HeaderV2      = 0x31313132
	HeaderVersion = HeaderV2
	Alignment     = 64
)

/** This is a component header - every entry in the CBFS
  will have this header.

  This is how the component is arranged in the ROM:

  --------------   <- 0
  component header
  --------------   <- sizeof(struct component)
  component name
  --------------   <- offset
  data
  ...
  --------------   <- offset + len
*/

const FileMagic = "LARCHIVE"

// This is kind of a mess. The file is aligned on 16 bytes. The size is 16 + Size. Why?
// Because in the beginning, IIRC, the AttrOffset and Offset weren't in there. Also the
// master record seems to be Type 2 but that's not documented.
// So we make a Tag, which is the thing we search on, and when we match, we read in the File,
// and, when we know what type it is, we pull that in. It's repeated work but it keeps it a
// bit simpler.
const TagSize = 16
type LarchiveTag struct {
	Magic      [8]byte
	Size       uint32
	Type       CBFSFileType
}

const FileSize = 24
type File struct {
	LarchiveTag
	AttrOffset uint32
	Offset     uint32
}

// The common fields of extended cbfs file attributes.
// Attributes are expected to start with tag/len, then append their
// specific fields.
type FileAttr struct {
	Tag  uint32
	Size uint32 // inclusize of Tag and Size
	Data []byte
}

type Tag uint32

const (
	Unused     Tag = 0
	Unused2        = 0xffffffff
	Compressed     = 0x42435a4c
	Hash           = 0x68736148
	PSCB           = 0x42435350
	ALCB           = 0x42434c41
)

type FileAttrCompression struct {
	Tag              Tag
	Size             uint32
	Compression      uint32
	DecompressedSize uint32
}

type FileAttrHash struct {
	Tag      Tag
	Size     uint32 // includes everything including data.
	HashType uint32
	Data     []byte
}

type FileAttrPos struct {
	Tag  Tag
	Size uint32 // includes everything including data.
	Pos  uint32
}

type FileAttrAlign struct {
	Tag   Tag
	Size  uint32 // includes everything including data.
	Align uint32
}

// Component sub-headers

// Following are component sub-headers for the "standard"
// component types

// this is the master cbfs header - it must be located somewhere available
// to bootblock (to load romstage). The last 4 bytes in the image contain its
// relative offset from the end of the image (as a 32-bit signed integer).
type CBFSHeader struct {
	File
	Magic         uint32
	Version       uint32
	RomSize       uint32
	BootBlockSize uint32
	Align         uint32 // always 64 bytes -- FOV
	Offset        uint32
	Architecture  CBFSArchitecture // integer, not name -- FOV
	_             uint32
}

type CBFSFile struct {
	File
	CBFSHeader
}

type CBFSArchitecture uint32

const (
	X86 CBFSArchitecture = 1
	ARM                  = 0x10
)

type Stage struct {
	File
	Compression uint32
	Entry       uint64
	LoadAddress uint64
	Size        uint32
	MemSize     uint32
}

type PayloadSegment struct {
	PayloadType uint32
	Compression uint32
	Offset      uint32
	LoadAddress uint64
	Size        uint32
	MemSize     uint32
}

type Payload struct {
	File
	Segs []PayloadSegment
}

// fix this mess later to use characters, not constants.
// I had done this once and it never made it into coreboot
// and I still don't know why.
type SegmentType uint32

const (
	SegCode   SegmentType = 0x434F4445
	SegData               = 0x44415441
	SegBSS                = 0x42535320
	SegParams             = 0x50415241
	SegEntry              = 0x454E5452
)

type OptionRom struct {
	File
	Compression uint32
	Size        uint32
}

// Each CBFS file type must implement at least this interface.
type CBFSReadWriter interface {
	String() string
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}
