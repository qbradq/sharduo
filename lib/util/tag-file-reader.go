package util

// TagFileReader reads objects from tag files.
type TagFileReader struct {
	ListFileReader
}

// ReadObject returns the next object in the file, or nil if no more objects are
// in the current file. Use HasErrors and Errors to see if there were errors.
func (f *TagFileReader) ReadObject() *TagFileObject {
	seg := f.ReadNextSegment()
	if seg == nil {
		return nil
	}
	tfo := NewTagFileObject()
	tfo.t = seg.Name
	for _, line := range seg.Contents {
		tfo.HandlePropertyLine(line)
	}
	if tfo.HasErrors() {
		f.errs = append(f.errs, tfo.Errors()...)
		return nil
	}
	return tfo
}
