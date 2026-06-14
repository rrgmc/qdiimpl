package qdii

type Context interface {
	MethodName() string
	ExecCount() int
	CallerFunc() string
	CallerFile() string
	CallerLine() int
	NotSupported()
}
