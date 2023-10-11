// Code generated by "qdiimpl"; DO NOT EDIT.
package main

import (
	"fmt"
	"io"
	"runtime"
)

type QDReaderContext struct {
	ExecCount  int
	CallerFunc string
	CallerFile string
	CallerLine int
}

type QDReader struct {
	execCount map[string]int
	implRead  func(debugCtx *QDReaderContext, p []byte) (n int, err error)
}

var _ io.Reader = (*QDReader)(nil)

type QDReaderOption func(*QDReader)

func NewQDReader(options ...QDReaderOption) *QDReader {
	ret := &QDReader{execCount: map[string]int{}}
	for _, opt := range options {
		opt(ret)
	}
	return ret
}

func (d *QDReader) Read(p []byte) (n int, err error) {
	return d.implRead(d.createContext("Read", d.implRead == nil), p)
}

func (d *QDReader) getCallerFuncName(skip int) (funcName string, file string, line int) {
	counter, file, line, success := runtime.Caller(skip)
	if !success {
		panic("runtime.Caller failed")
	}
	return runtime.FuncForPC(counter).Name(), file, line
}

func (d *QDReader) checkCallMethod(methodName string, implIsNil bool) (count int) {
	if implIsNil {
		panic(fmt.Errorf("[QDReader] method '%s' not implemented", methodName))
	}
	d.execCount[methodName]++
	return d.execCount[methodName]
}

func (d *QDReader) createContext(methodName string, implIsNil bool) *QDReaderContext {
	callerFunc, callerFile, callerLine := d.getCallerFuncName(3)
	return &QDReaderContext{ExecCount: d.checkCallMethod(methodName, implIsNil), CallerFunc: callerFunc, CallerFile: callerFile, CallerLine: callerLine}
}

// Options

func WithQDReaderRead(implRead func(debugCtx *QDReaderContext, p []byte) (n int, err error)) QDReaderOption {
	return func(d *QDReader) {
		d.implRead = implRead
	}
}
