package gaefs

import (
	"errors"
	"fmt"
	"os"
)

var pf func(format string, a ...interface{}) (int, error) = fmt.Printf
var pfRestore func(format string, a ...interface{}) (int, error) = fmt.Printf
var spf func(format string, a ...interface{}) string = fmt.Sprintf

var (
	tdir       string = "fsd" // FileSys filesystem directory
	tdirsep    string = tdir + ","
	tfil       string = "fsf" // FileSys filesystem file
	dummdidumm        = ""
)

const sep = "/" // no, package path does not provide it; yes, we do need it.

var (
	ErrFileClosed        = errors.New("File is closed")
	ErrOutOfRange        = errors.New("Out of range")
	ErrTooLarge          = errors.New("Too large")
	ErrFileNotFound      = os.ErrNotExist
	ErrFileExists        = os.ErrExist
	ErrDestinationExists = os.ErrExist
)
