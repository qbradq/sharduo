package util

import (
	"bufio"
	"fmt"
	"io"
)

// TagFileReader reads objects from tag files.
type TagFileReader struct {
	s          bufio.Scanner
	ln         int
	nextObject *TagFileObject
	eof        bool
	errs       []error
	f          *SerializeableFactory
}

// NewTagFileReader returns a new TagFileReader object ready to use for input.
func NewTagFileReader(r io.Reader, f *SerializeableFactory) *TagFileReader {
	return &TagFileReader{
		s: *bufio.NewScanner(r),
		f: f,
	}
}

// handleError processes errors for the file reader.
func (f *TagFileReader) handleError(err error) {
	f.errs = append(f.errs, err)
}

// Errors returns the slice of all accumulated errors.
func (f *TagFileReader) Errors() []error {
	return f.errs
}

// CreateObject creates and deserializes a new object from a TagFileObject.
func (f *TagFileReader) CreateObject() Serializeable {
	s := f.f.New(f.nextObject.t, f.nextObject)
	if s == nil {
		f.handleError(fmt.Errorf("unable to create object of type %s", f.nextObject.t))
	} else {
		s.Deserialize(f.nextObject)
	}
	f.nextObject = NewTagFileObject(f)
	return s
}

// ReadObject returns the next object in the file, or nil, io.EOF on end of
// file or end of data store in the current file. nil, nil is returned when
// there was an error creating or deserializing the object. Use Errors() to
// inspect the accumulated errors. error will only ever be nil or io.EOF.
func (f *TagFileReader) ReadObject() (Serializeable, error) {
	if f.eof {
		return nil, io.EOF
	}

	// Process property lines until we hit the next object definition
	for {
		f.ln++
		if !f.s.Scan() {
			// Last object in file
			if f.s.Err() == nil {
				f.eof = true
				if f.nextObject == nil {
					return nil, io.EOF
				} else {
					return f.CreateObject(), nil
				}
			}
			// Genuine error
			f.handleError(fmt.Errorf("error loading tag file at line %d:%w", f.ln, f.s.Err()))
			return nil, nil
		}
		line := stripLine(f.s.Text())

		// Comment or empty line, ignore
		if isCommentLine(line) {
			continue
		}

		// First object tag of the tag file
		if f.nextObject == nil {
			if !isTypeLine(line) {
				f.handleError(fmt.Errorf("error loading tag file at line %d:initial object type not found", f.ln))
				return nil, nil
			}
			f.nextObject = NewTagFileObject(f)
			f.nextObject.t = extractObjectType(line)
		} else if isTypeLine(line) {
			f.nextObject.t = extractObjectType(line)
			s := f.CreateObject()
			f.nextObject = NewTagFileObject(f)
			if err := f.nextObject.HandleTypeLine(line); err != nil {
				f.handleError(fmt.Errorf("error deserializing object at line %d:%w", f.ln, f.s.Err()))
				return nil, nil
			}
			return s, nil
		}

		// Handle property line for the next object and go to the next line
		if err := f.nextObject.HandlePropertyLine(line); err != nil {
			f.handleError(fmt.Errorf("error loading tag file at line %d:%w", f.ln, err))
			return nil, nil
		}
	}
}
