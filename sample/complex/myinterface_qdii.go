// Code generated by "qdiimpl"; DO NOT EDIT.
package main

import (
	"context"
	"fmt"
	"runtime"
)

type QDMyInterfaceContext struct {
	ExecCount  int
	CallerFunc string
	CallerFile string
	CallerLine int
	Data       any
}

type QDMyInterface[T any, X II] struct {
	DataQDII any

	execCount       map[string]int
	implCloseNotify func(debugCtx *QDMyInterfaceContext) <-chan bool
	implData        func(debugCtx *QDMyInterfaceContext)
	implGet         func(debugCtx *QDMyInterfaceContext, ctx context.Context, name string) (T, error)
	implOther       func(debugCtx *QDMyInterfaceContext, si SecondInterface) int
	implOther2      func(debugCtx *QDMyInterfaceContext, ti ThirdInterface[T]) int
	implSet         func(debugCtx *QDMyInterfaceContext, ctx context.Context, name string, value T) error
	implUnnamed     func(debugCtx *QDMyInterfaceContext, p0 bool, p1 string)
	implXGet        func(debugCtx *QDMyInterfaceContext, ss *SI) *SI
	implinternal    func(debugCtx *QDMyInterfaceContext) bool
}

type QDMyInterfaceOption[T any, X II] func(*QDMyInterface[T, X])

func NewQDMyInterface[T any, X II](options ...QDMyInterfaceOption[T, X]) *QDMyInterface[T, X] {
	ret := &QDMyInterface[T, X]{execCount: map[string]int{}}
	for _, opt := range options {
		opt(ret)
	}
	return ret
}

func (d *QDMyInterface[T, X]) CloseNotify() <-chan bool {
	return d.implCloseNotify(d.createContext("CloseNotify", d.implCloseNotify == nil))
}

func (d *QDMyInterface[T, X]) Data() {
	d.implData(d.createContext("Data", d.implData == nil))
}

func (d *QDMyInterface[T, X]) Get(ctx context.Context, name string) (T, error) {
	return d.implGet(d.createContext("Get", d.implGet == nil), ctx, name)
}

func (d *QDMyInterface[T, X]) Other(si SecondInterface) int {
	return d.implOther(d.createContext("Other", d.implOther == nil), si)
}

func (d *QDMyInterface[T, X]) Other2(ti ThirdInterface[T]) int {
	return d.implOther2(d.createContext("Other2", d.implOther2 == nil), ti)
}

func (d *QDMyInterface[T, X]) Set(ctx context.Context, name string, value T) error {
	return d.implSet(d.createContext("Set", d.implSet == nil), ctx, name, value)
}

func (d *QDMyInterface[T, X]) Unnamed(p0 bool, p1 string) {
	d.implUnnamed(d.createContext("Unnamed", d.implUnnamed == nil), p0, p1)
}

func (d *QDMyInterface[T, X]) XGet(ss *SI) *SI {
	return d.implXGet(d.createContext("XGet", d.implXGet == nil), ss)
}

func (d *QDMyInterface[T, X]) internal() bool {
	return d.implinternal(d.createContext("internal", d.implinternal == nil))
}

func (d *QDMyInterface[T, X]) getCallerFuncName(skip int) (funcName string, file string, line int) {
	counter, file, line, success := runtime.Caller(skip)
	if !success {
		panic("runtime.Caller failed")
	}
	return runtime.FuncForPC(counter).Name(), file, line
}

func (d *QDMyInterface[T, X]) checkCallMethod(methodName string, implIsNil bool) (count int) {
	if implIsNil {
		panic(fmt.Errorf("[QDMyInterface] method '%s' not implemented", methodName))
	}
	d.execCount[methodName]++
	return d.execCount[methodName]
}

func (d *QDMyInterface[T, X]) createContext(methodName string, implIsNil bool) *QDMyInterfaceContext {
	callerFunc, callerFile, callerLine := d.getCallerFuncName(3)
	return &QDMyInterfaceContext{ExecCount: d.checkCallMethod(methodName, implIsNil), CallerFunc: callerFunc, CallerFile: callerFile, CallerLine: callerLine, Data: d.DataQDII}
}

// Options

func WithQDMyInterfaceDataQDII[T any, X II](data any) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.DataQDII = data
	}
}

func WithQDMyInterfaceCloseNotify[T any, X II](implCloseNotify func(debugCtx *QDMyInterfaceContext) <-chan bool) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implCloseNotify = implCloseNotify
	}
}

func WithQDMyInterfaceData[T any, X II](implData func(debugCtx *QDMyInterfaceContext)) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implData = implData
	}
}

func WithQDMyInterfaceGet[T any, X II](implGet func(debugCtx *QDMyInterfaceContext, ctx context.Context, name string) (T, error)) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implGet = implGet
	}
}

func WithQDMyInterfaceOther[T any, X II](implOther func(debugCtx *QDMyInterfaceContext, si SecondInterface) int) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implOther = implOther
	}
}

func WithQDMyInterfaceOther2[T any, X II](implOther2 func(debugCtx *QDMyInterfaceContext, ti ThirdInterface[T]) int) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implOther2 = implOther2
	}
}

func WithQDMyInterfaceSet[T any, X II](implSet func(debugCtx *QDMyInterfaceContext, ctx context.Context, name string, value T) error) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implSet = implSet
	}
}

func WithQDMyInterfaceUnnamed[T any, X II](implUnnamed func(debugCtx *QDMyInterfaceContext, p0 bool, p1 string)) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implUnnamed = implUnnamed
	}
}

func WithQDMyInterfaceXGet[T any, X II](implXGet func(debugCtx *QDMyInterfaceContext, ss *SI) *SI) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implXGet = implXGet
	}
}

func WithQDMyInterfaceinternal[T any, X II](implinternal func(debugCtx *QDMyInterfaceContext) bool) QDMyInterfaceOption[T, X] {
	return func(d *QDMyInterface[T, X]) {
		d.implinternal = implinternal
	}
}
