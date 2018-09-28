package cbfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/linuxboot/fiano/pkg/fmap"
)

type SegReader struct {
	T FileType
	F func(r CountingReader, f *File) (ReadWriter, error)
	N string
}

var SegReaders = make(map[FileType]*SegReader)

func RegisterFileReader(f *SegReader) error {
	if r, ok := SegReaders[f.T]; ok {
		return fmt.Errorf("RegisterFileType: Slot of %v is owned by %s, can't add %s", r.T, r.N, f.N)
	}
	SegReaders[f.T] = f
	Debug("Registered %v", f)
	return nil
}

func NewImage(in io.ReadSeeker) (*Image, error) {
	// Suck the image in. Todo: write a thing that implements
	// ReadSeeker on a []byte.
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	if _, err := in.Seek(0, 0); err != nil {
		return nil, err
	}
	f, m, err := fmap.Read(in)
	if err != nil {
		return nil, err
	}
	Debug("Fmap %v", f)
	var i = &Image{Offset: -1, FMAP: f, FMAPMetadata: m, Data: b}
	var x = int(-1)
	for i, a := range f.Areas {
		Debug("Check %v", a.Name.String())
		if a.Name.String() == "COREBOOT" {
			x = i
			break
		}
	}
	if x == -1 {
		return nil, fmt.Errorf("No CBFS in fmap")
	}
	i.Index = x
	Debug("COREBOOT is the %d entry", x)
	fr, err := f.ReadArea(in, x)
	if err != nil {
		return nil, err
	}
	r := NewCountingReader(fr)

	for {
		var f File
		var m Magic
		err := Align(r)
		if err == io.EOF {
			return i, nil
		}
		if err != nil {
			return nil, err
		}
		recStart := r.Count()
		err = Read(r, m[:])
		if err == io.EOF {
			return i, nil
		}
		if err != nil {
			return nil, err
		}
		if string(m[:]) != FileMagic {
			continue
		}
		Debug("It is an LARCHIVE at %#x", int(r.Count())-len(FileMagic))
		if i.Offset < 0 {
			i.Offset = int(recStart)
		}
		if err := Read(r, &f.FileHeader); err != nil {
			Debug("Reading the File failed: %v", err)
			return nil, err
		}
		f.RomOffset = recStart
		Debug("It is %v type %v", f, f.Type)
		sr, ok := SegReaders[f.Type]
		if !ok {
			return nil, fmt.Errorf("%v: unknown type %v", f, f.Type)
		}
		headSize := r.Count() - recStart
		Debug("Namelen %d %d ", f.SubHeaderOffset, f.SubHeaderOffset-headSize)
		n, err := ReadName(r, &f, f.SubHeaderOffset-headSize)
		if err != nil {
			return nil, err
		}
		f.Name = n
		Debug("Count after name is %#x", r.Count())
		Debug("Found a SegReader for this %d size section: %v", f.Size, n)
		s, err := sr.F(r, &f)
		if err != nil {
			return nil, err
		}
		Debug("Segment was readable")
		i.Segs = append(i.Segs, s)
		Debug("r.Count is now %#x", r.Count())
	}
	return i, nil
}

func (i *Image) WriteFile(name string, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return err
	}
	// do the simple thing, fill the file with 0xff and then write the things 
	ff := make([]byte, i.FMAP.Size)
	for i := range ff {
		ff[i] = 0xff
	}
	if _, err := f.WriteAt(ff, 0); err != nil {
		return err
	}
	if err := fmap.Write(f, i.FMAP, i.FMAPMetadata); err != nil {
		return err
	}
	for _, a := range i.FMAP.Areas {
		if _, err := f.WriteAt(i.Data[a.Offset:a.Size], int64(a.Offset)); err != nil {
			return err
		}
	}
	Debug("Wrote the fmap")
	return nil
}

func (i *Image) String() string {
	var s = "FMAP REGION: COREBOOT\nName\t\t\t\tOffset\tType\t\tSize\tComp\n"
	for _, seg := range i.Segs {
		s = s + seg.String() + "\n"
	}
	return s
}
