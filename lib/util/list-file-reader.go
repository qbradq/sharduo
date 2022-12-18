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
	nextSegmentName string
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

// readNextLine returns the next line of the file and no error, or io.EOF when
// there are no more lines. Other errors may be returned as well.
func (f *ListFileReader) readNextLine() (string, error) {
	if f.sawEOF {
		return "", io.EOF
	}
	if !f.scanner.Scan() {
		err := f.scanner.Err()
		if err != nil {
			f.errs = append(f.errs, err)
			return "", err
		}
		f.sawEOF = true
		return "", io.EOF
	}
	return strings.TrimSpace(f.scanner.Text()), nil
}

// StreamNextSegmentHeader continues reading the list file until the next non-
// empty, non-comment line. It expects this line to be a segment header in the
// form [SEGMENT_NAME] . The name of the segment is returned. The empty string
// means end of file or error. Use HasErrors and Errors to inspect the error
// state.
func (f *ListFileReader) StreamNextSegmentHeader() string {
	var line string
	var err error
	if f.nextSegmentName != "" {
		ret := f.nextSegmentName
		f.nextSegmentName = ""
		return ret
	}

	for {
		line, err = f.readNextLine()
		if err != nil {
			return ""
		}
		if len(line) == 0 || f.isCommentLine(line) {
			continue
		}
		if f.isSegmentLine(line) {
			return f.extractSegmentName(line)
		}
		f.errs = append(f.errs, errors.New("expected a segment line, found a list entry"))
		return ""
	}
}

// StreamNextEntry continues reading the list file until the next non-empty,
// non-comment line. It expects this line to be an entry line (not a segment
// header). The empty string return value means no more entries in this segment,
// or an error condition. Use HasErrors and Errors to inspect the error state.
func (f *ListFileReader) StreamNextEntry() string {
	var line string
	var err error

	for {
		line, err = f.readNextLine()
		if err != nil {
			return ""
		}
		if len(line) == 0 || f.isCommentLine(line) {
			continue
		}
		if f.isSegmentLine(line) {
			f.nextSegmentName = f.extractSegmentName(line)
			return ""
		}
		return line
	}
}

// SkipCurrentSegment runs the file forward to the start of the next segment
// header. StreamNextSegment and ReadNextSegment are both sane to call after
// a call to SkipCurrentSegment. On error or end of file this function returns
// false. The error state can be inspected after this call with HasErrors and
// Errors.
func (f *ListFileReader) SkipCurrentSegment() bool {
	var line string
	var err error
	for {
		line, err = f.readNextLine()
		if err != nil {
			return false
		}
		if len(line) > 0 && f.isSegmentLine(line) {
			f.nextSegmentName = f.extractSegmentName(line)
			return true
		}
		// Entry line, ignore
	}
}

// ReadNextSegment returns the next segment in the current reader stream or nil
// if the end of the stream has been reached. Use HasErrors and Errors to
// see if there were errors in execution.
func (f *ListFileReader) ReadNextSegment() *ListFileSegment {
	lfs := &ListFileSegment{}
	lfs.Name = f.StreamNextSegmentHeader()
	if lfs.Name == "" {
		return nil
	}
	for {
		e := f.StreamNextEntry()
		if e == "" {
			break
		}
		lfs.Contents = append(lfs.Contents, e)
	}
	return lfs
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
