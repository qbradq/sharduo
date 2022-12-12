package util

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// ListFileSegment represents one segment of the list file
type ListFileSegment struct {
	Name     string
	Contents []string
}

// ListFileReader reads ini-like list files
type ListFileReader struct {
	errs            []error
	scanner         *bufio.Scanner
	nextFileSegment *ListFileSegment
	sawEOF          bool
}

func (f *ListFileReader) isCommentLine(line string) bool {
	return len(line) == 0 || line[0] == ';'
}

func (f *ListFileReader) isSegmentLine(line string) bool {
	return len(line) > 1 && line[0] == '[' && line[len(line)-1] == ']'
}

func (f *ListFileReader) extractSegmentName(line string) string {
	if len(line) < 3 {
		f.errs = append(f.errs, errors.New("empty segment header"))
		return ""
	}
	return line[1 : len(line)-1]
}

// HasErrors returns true if there are any errors
func (f *ListFileReader) HasErrors() bool {
	return len(f.errs) > 0
}

// Errors returns the slice of all accumulated errors
func (f *ListFileReader) Errors() []error {
	return f.errs
}

// StartReading should be called to begin reading from a new reader
func (f *ListFileReader) StartReading(r io.Reader) {
	f.scanner = bufio.NewScanner(r)
	f.nextFileSegment = &ListFileSegment{}
	f.sawEOF = false
}

// ReadNextSegment returns the next segment in the current reader stream or nil
// if the end of the stream has been reached.
func (f *ListFileReader) ReadNextSegment() *ListFileSegment {
	for {
		if f.sawEOF {
			return nil
		}
		if !f.scanner.Scan() {
			err := f.scanner.Err()
			if err != nil {
				f.errs = append(f.errs, err)
				return nil
			}
			f.sawEOF = true
			return f.nextFileSegment
		}
		line := strings.TrimSpace(f.scanner.Text())
		if f.isCommentLine(line) {
			// Do nothing
		} else if f.isSegmentLine(line) {
			ret := f.nextFileSegment
			f.nextFileSegment = &ListFileSegment{
				Name: f.extractSegmentName(line),
			}
			// Special case, if the root segment has no lines do not emmit it
			if ret.Name == "" && len(ret.Contents) == 0 {
				continue
			}
			return ret
		} else {
			f.nextFileSegment.Contents = append(f.nextFileSegment.Contents, line)
		}
	}
}

// ReadSegments returns all of the list segments from the reader. Use HasErrors
// and Errors to look for error conditions found within the file.
func (f *ListFileReader) ReadSegments(r io.Reader) []*ListFileSegment {
	var ret []*ListFileSegment

	f.StartReading(r)
	for {
		s := f.ReadNextSegment()
		if s != nil {
			ret = append(ret, s)
		}
		if s == nil {
			break
		}
	}
	return ret
}
