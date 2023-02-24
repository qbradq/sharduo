package util

var templateObjectGetter func(string) *TagFileObject

// SetTemplateObjectGetter sets the function used to get template objects by
// name.
func SetTemplateObjectGetter(fn func(string) *TagFileObject) {
	templateObjectGetter = fn
}
