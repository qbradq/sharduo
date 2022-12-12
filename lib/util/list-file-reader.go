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
	errs []error
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

// ReadSegments returns all of the list segments from the reader. Use HasErrors
// and Errors to look for error conditions found within the file.
func (f *ListFileReader) ReadSegments(r io.Reader) []*ListFileSegment {
	var ret []*ListFileSegment

	scanner := bufio.NewScanner(r)
	nextFileSegment := &ListFileSegment{}
	for {
		if !scanner.Scan() {
			err := scanner.Err()
			if !errors.Is(err, io.EOF) {
				f.errs = append(f.errs, err)
			}
			return append(ret, nextFileSegment)
		}
		line := strings.TrimSpace(scanner.Text())
		if f.isCommentLine(line) {
			// Do nothing
		} else if f.isSegmentLine(line) {
			ret = append(ret, nextFileSegment)
			nextFileSegment = &ListFileSegment{
				Name: f.extractSegmentName(line),
			}
		} else {
			nextFileSegment.Contents = append(nextFileSegment.Contents, line)
		}
	}
}
