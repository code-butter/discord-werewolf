package lib

import (
	"log"
	"os"

	"github.com/pkg/errors"
)

func Fatal(err interface{}) {
	log.Printf("%+v\n", err)
	os.Exit(1)
}

type UnwrappableError interface {
	Unwrap() []error
}

type StackTracer interface {
	Error() string
	StackTrace() errors.StackTrace
}
