package cbfs

import (
	"io"
	"log"
)

func init() {
	if err := RegisterFileReader(&SegReader{Type: TypeSPD, Name: "spd", New: NewSPD}); err != nil {
		log.Fatal(err)
	}
}

//NewSPD returns a ReadWriter interface for the CBFS type TypeSPD
func NewSPD(f *File) (ReadWriter, error) {
	rec := &SPDRecord{File: *f}
	return rec, nil
}

func (r *SPDRecord) Read(in io.ReadSeeker) error {
	return nil
}

func (r *SPDRecord) String() string {
	return recString(r.File.Name, r.RecordStart, r.Type.String(), r.Size, "none")
}

func (r *SPDRecord) Write(w io.Writer) error {
	return Write(w, r.FData)
}

//Header returns a pointer to the corresponding File
func (r *SPDRecord) Header() *File {
	return &r.File
}
