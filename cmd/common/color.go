package common

import (
	"github.com/fatih/color"
	"io"
)

type CWriter struct {
	Writer io.Writer
}

func (writer *CWriter) Write(c color.Attribute, data ...interface{}) {
	cw := color.New(c)
	cw.Fprint(writer.Writer, data...)
}

func (writer *CWriter) WriteF(c color.Attribute, format string, data ...interface{}) {
	cw := color.New(c)
	cw.Fprintf(writer.Writer, format, data...)
}


