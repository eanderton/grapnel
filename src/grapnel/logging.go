package grapnel

import (
  log "github.com/ngmoco/timber"
)

var loggingConfigured bool = false
var LoggingVerbose bool = false
var LoggingQuiet bool = false

func initTestLogging() {
  if loggingConfigured {
    return
  }

  log.AddLogger(log.ConfigLogger{
    LogWriter: new(log.ConsoleWriter),
    Level:     log.DEBUG,
    Formatter: log.NewPatFormatter("[%L] %M"),
  })

  loggingConfigured = true
}


func InitLogging() {
  if loggingConfigured {
    return
  }

  // configure logging level based on flags
  var logLevel log.Level
  if LoggingQuiet {
    logLevel = log.ERROR
  } else if LoggingVerbose {
    logLevel = log.INFO
  } else {
    logLevel = log.WARNING
  }
  // set a vanilla console writer
  log.AddLogger(log.ConfigLogger{
    LogWriter: new(log.ConsoleWriter),
    Level:     logLevel,
    Formatter: log.NewPatFormatter("[%L] %M"),
  })

  loggingConfigured = true
}
