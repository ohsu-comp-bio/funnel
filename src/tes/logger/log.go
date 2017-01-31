package logger

import (
  "io"
  "github.com/Sirupsen/logrus"
)

func init() {
  logrus.SetFormatter(&TextFormatter{
    DisableTimestamp: true,
  })
}

type Logger interface {
  Debug(string, ...interface{})
  Info(string, ...interface{})
  Error(string, ...interface{})
  SetOutput(io.Writer)
}

func New(ns string, args ...interface{}) Logger {
  f := fields(args...)
  f["ns"] = ns
  return &logger{f}
}

type logger struct {
  fields map[string]interface{}
}

func fields(args ...interface{}) map[string]interface{} {
  f := make(map[string]interface{}, len(args) / 2)
  for i := 0; i < len(args); i += 2 {
    k := args[i].(string)
    v := args[i + 1]
    f[k] = v
  }
  if len(args) % 2 != 0 {
    f["unknown"] = args[len(args) - 1]
  }
  return f
}

func (l *logger) Debug(msg string, args ...interface{}) {
  f := fields(args...)
  logrus.WithFields(l.fields).WithFields(f).Debug(msg)
}

func (l *logger) Info(msg string, args ...interface{}) {
  f := fields(args...)
  logrus.WithFields(l.fields).WithFields(f).Info(msg)
}

func (l *logger) Error(msg string, args ...interface{}) {
  var f map[string]interface{}
  if len(args) == 1 {
    f = fields("error", args[0])
  } else {
    f = fields(args...)
  }
  logrus.WithFields(l.fields).WithFields(f).Error(msg)
}

func (l *logger) SetOutput(w io.Writer) {
  logrus.SetOutput(w)
}

var rootLogger = New("tes")

func Debug(msg string, args ...interface{}) {
  rootLogger.Debug(msg, args...)
}
func Info(msg string, args ...interface{}) {
  rootLogger.Info(msg, args...)
}
func Error(msg string, args ...interface{}) {
  rootLogger.Error(msg, args...)
}
