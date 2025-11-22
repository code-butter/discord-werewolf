package lib

import (
	"log"
	"os"
)

func Fatal(err interface{}) {
	log.Printf("%+v\n", err)
	os.Exit(1)
}
