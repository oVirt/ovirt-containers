/*
Copyright (c) 2017 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file contains functions useful for generating log files.

package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// The log file.
//
var path string
var file *os.File

// The writers for the different levels.
//
var errorWriter io.Writer
var infoWriter io.Writer
var debugWriter io.Writer

// prefixWriter implements a writer that adds a prefix to each line.
//
type prefixWriter struct {
	prefix string
	stream io.Writer
	start  bool
}

// newPrefixWriter creates a new prefix writer that adds the given
// prefix and writes the modified lines to the given stream.
//
func newPrefixWriter(stream io.Writer, prefix string) io.Writer {
	p := new(prefixWriter)
	p.prefix = prefix
	p.stream = stream
	p.start = true
	return p
}

func (p *prefixWriter) Write(data []byte) (count int, err error) {
	buffer := new(bytes.Buffer)
	buffer.Grow(len(data))
	for _, char := range data {
		if p.start {
			buffer.WriteString(p.prefix)
			p.start = false
		}
		buffer.WriteByte(char)
		if char == 10 {
			p.start = true
		}
	}
	count = len(data)
	_, err = p.stream.Write(buffer.Bytes())
	return
}

// newLogWriter creates a log writer that will write to the given file
// and console. Each line will have the prefix added, in the given
// color. The file and the console can be nil.
//
func newLogWriter(file, console io.Writer, prefix, color string) io.Writer {
	writers := make([]io.Writer, 0)
	if file != nil {
		plain := fmt.Sprintf("[%s] ", prefix)
		file = newPrefixWriter(file, plain)
		writers = append(writers, file)
	}
	if console != nil {
		colored := fmt.Sprintf("\033[%sm[%s]\033[m ", color, prefix)
		console = newPrefixWriter(console, colored)
		writers = append(writers, console)
	}
	return io.MultiWriter(writers...)
}

// Open creates a log file with the given name and the .log extension,
// and configures the log to write to it.
//
func Open(name string) error {
	var err error

	path, err = filepath.Abs(name + ".log")
	if err != nil {
		return err
	}
	file, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	errorWriter = newLogWriter(file, os.Stderr, "ERROR", "0;31")
	infoWriter = newLogWriter(file, os.Stdout, "INFO", "0;32")
	debugWriter = newLogWriter(file, nil, "DEBUG", "0;34")
	return nil
}

// Close closes the log file.
//
func Close() error {
	return file.Close()
}

// Path the returns the path of the log file.
//
func Path() string {
	return path
}

// Info sends an informative message to the log file and to the standard
// ouptut of the process.
//
func Info(format string, args ...interface{}) {
	write(infoWriter, format, args...)
}

// Debug sends a debug message to the log file.
//
func Debug(format string, args ...interface{}) {
	write(debugWriter, format, args...)
}

// Error sends an error message to the log file and to the standard
// error stream of the process.
//
func Error(format string, args ...interface{}) {
	write(errorWriter, format, args...)
}

func write(writer io.Writer, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writer.Write([]byte(message))
	writer.Write([]byte("\n"))
}

// InfoWriter returns the writer that writes informative messages to the
// log file and to the standard output of the process.
//
func InfoWriter() io.Writer {
	return infoWriter
}

// DebugWriter returns the writer that writes debug messages to the log
// file.
//
func DebugWriter() io.Writer {
	return debugWriter
}

// ErrorWriter returns the writer that writes error messages to the log
// file and to the standard error stream of the process.
// file.
//
func ErrorWriter() io.Writer {
	return errorWriter
}
