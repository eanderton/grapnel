package log
/*
Copyright (c) 2014 Eric Anderton <eric.t.anderton@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
  "log"
)

const (
  DEBUG = iota
  INFO
  WARN
  ERROR
  FATAL
)

var (
  logLevel int = WARN
)

// Set the log level for the entire application
func SetGlobalLogLevel(level int) {
  logLevel = level
}

func SetFlags(flags int) {
  log.SetFlags(flags)
}

// A logger that contains a prefix
type Logger struct {
  prefix []string
}

func NewLogger() *Logger {
  return &Logger{}
}

func (self *Logger) AddPrefix(prefix string) *Logger {
  self.prefix = append(self.prefix, prefix)
  return self
}

func (self *Logger) log(level int, args... interface{}) {
  if len(args) == 0 { return }

  // determine if we're a format style print or not
  fmtString, usePrintf := args[0].(string)

  // handle fatal events specially
  if level == FATAL {
    if usePrintf {
      log.Fatalf(fmtString, args[1:]...)
    } else {
      log.Fatal(args...)
    }
  } else {
    if usePrintf {
      log.Printf(fmtString, args[1:]...)
    } else {
      log.Print(args...)
    }
  }
}

func (self *Logger) Debug(args... interface{}) {
  self.log(DEBUG, args...)
}

func (self *Logger) Info(args... interface{}) {
  self.log(INFO, args...)
}

func (self *Logger) Warn(args... interface{}) {
  self.log(WARN, args...)
}

func (self *Logger) Error(args... interface{}) {
  self.log(ERROR, args...)
}

func (self *Logger) Fatal(args... interface{}) {
  self.log(FATAL, args...)
}

var RootLogger *Logger = NewLogger()

func Debug(args... interface{}) {
  RootLogger.log(DEBUG, args...)
}

func Info(args... interface{}) {
  RootLogger.log(INFO, args...)
}

func Warn(args... interface{}) {
  RootLogger.log(WARN, args...)
}

func Error(args... interface{}) {
  RootLogger.log(ERROR, args...)
}

func Fatal(args... interface{}) {
  RootLogger.log(FATAL, args...)
}

