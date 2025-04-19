package logger

import (
	"log"
	"os"
)

var logg = log.New(os.Stdout, "[yubin-smtp-server]:", log.LstdFlags)

func Println(v ...any) {
	logg.Println(v...)
}
