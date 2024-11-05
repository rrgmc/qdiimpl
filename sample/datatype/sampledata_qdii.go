// Code generated by "qdiimpl"; DO NOT EDIT.
package main

import (
	"fmt"
	idata "github.com/rrgmc/qdiimpl/sample/datatype/idata"
	"runtime"
	"sync"
)

type QDSampleDataContext struct {
	MethodName     string
	ExecCount      int
	CallerFunc     string
	CallerFile     string
	CallerLine     int
	isNotSupported bool
	Data           *idata.IData
}

// NotSupported should be called if the current callback don't support the passed arguments.
// The function return values will be ignored.
func (c *QDSampleDataContext) NotSupported() {
	c.isNotSupported = true
}

type QDSampleData struct {
	Data *idata.IData

	lock                   sync.Mutex
	execCount              map[string]int
	fallback               SampleData
	onMethodNotImplemented func(qdCtx *QDSampleDataContext, hasCallbacks bool) error
	implGet                []func(qdCtx *QDSampleDataContext, name string) (any, error)
}

var _ SampleData = (*QDSampleData)(nil)

type QDSampleDataOption func(*QDSampleData)

func NewQDSampleData(options ...QDSampleDataOption) *QDSampleData {
	ret := &QDSampleData{execCount: map[string]int{}}
	for _, opt := range options {
		opt(ret)
	}
	return ret
}

// Get implements [main.SampleData.Get].
func (d *QDSampleData) Get(name string) (any, error) {
	const methodName = "Get"
	for _, impl := range d.implGet {
		qctx := d.createContext(methodName)
		r0, r1 := impl(qctx, name)
		if !qctx.isNotSupported {
			d.addCallMethod(methodName)
			return r0, r1
		}
	}
	if d.fallback != nil {
		return d.fallback.Get(name)
	}
	panic(d.methodNotImplemented(d.createContext(methodName), len(d.implGet) > 0))
}

func (d *QDSampleData) getCallerFuncName(skip int) (funcName string, file string, line int) {
	counter, file, line, success := runtime.Caller(skip)
	if !success {
		panic("runtime.Caller failed")
	}
	return runtime.FuncForPC(counter).Name(), file, line
}

func (d *QDSampleData) addCallMethod(methodName string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.execCount[methodName]++
}

func (d *QDSampleData) createContext(methodName string) *QDSampleDataContext {
	callerFunc, callerFile, callerLine := d.getCallerFuncName(3)
	d.lock.Lock()
	defer d.lock.Unlock()
	return &QDSampleDataContext{
		MethodName: methodName,
		ExecCount:  d.execCount[methodName],
		CallerFunc: callerFunc,
		CallerFile: callerFile,
		CallerLine: callerLine,
		Data:       d.Data,
	}
}

func (d *QDSampleData) methodNotImplemented(qdCtx *QDSampleDataContext, hasCallbacks bool) error {
	if d.onMethodNotImplemented != nil {
		if err := d.onMethodNotImplemented(qdCtx, hasCallbacks); err != nil {
			return err
		}
	}
	msg := "not implemented"
	if hasCallbacks {
		msg = "not supported by any callbacks"
	}
	msg = "not supported by any callbacks"
	return fmt.Errorf("[QDSampleData] method '%s' %s", qdCtx.MethodName, msg)
}

// Options

func WithQDData(data *idata.IData) QDSampleDataOption {
	return func(d *QDSampleData) {
		d.Data = data
	}
}

func WithQDFallback(fallback SampleData) QDSampleDataOption {
	return func(d *QDSampleData) {
		d.fallback = fallback
	}
}

func WithQDOnMethodNotImplemented(m func(qdCtx *QDSampleDataContext, hasCallbacks bool) error) QDSampleDataOption {
	return func(d *QDSampleData) {
		d.onMethodNotImplemented = m
	}
}

// WithQDGet implements [main.SampleData.Get].
func WithQDGet(implGet func(qdCtx *QDSampleDataContext, name string) (any, error)) QDSampleDataOption {
	return func(d *QDSampleData) {
		d.implGet = append(d.implGet, implGet)
	}
}
