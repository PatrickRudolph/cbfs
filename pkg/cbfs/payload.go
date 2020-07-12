package cbfs

import (
	"fmt"
	"io"
	"log"
)

func init() {
	if err := RegisterFileReader(&SegReader{Type: TypeSELF, Name: "Payload", New: NewPayloadRecord}); err != nil {
		log.Fatal(err)
	}
}

func NewPayloadRecord(f *File) (ReadWriter, error) {
	p := &PayloadRecord{File: *f, Self: PayloadSELFRecord{File: *f}}
	return p, nil
}

func (p *PayloadRecord) Read(in io.ReadSeeker) error {
	if p.File.Type == TypeSELF {
		return p.Self.Read(in)
	} else {
		return fmt.Errorf("Unsupported CBFS file type: #%x", p.File.Type)
	}
}

func (h *PayloadRecord) String() string {
	if h.File.Type == TypeSELF {
		return h.Self.String()
	} else {
		return ""
	}
}

func (r *PayloadRecord) Write(w io.Writer) error {
	if r.File.Type == TypeSELF {
		return r.Self.Write(w)
	} else {
		return fmt.Errorf("Unsupported CBFS file type: #%x", r.File.Type)
	}
}

func (r *PayloadRecord) Header() *File {
	if r.File.Type == TypeSELF {
		return r.Self.Header()
	} else {
		return nil
	}
}
