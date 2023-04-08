package marshal

// insertFunction inserts the object into the datastore
var insertFunction func(interface{})

// SetInsertFunction sets the function that inserts objects into the datastore
func SetInsertFunction(fn func(interface{})) {
	insertFunction = fn
}
