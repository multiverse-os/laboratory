package output

import "io"

// Writer can write using a given io stream
type Writer interface {
	Write(io.Writer) error
}
