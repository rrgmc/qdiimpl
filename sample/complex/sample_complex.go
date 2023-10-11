package main

import "context"

type II interface {
	Setme()
}

type SI struct {
	A int
}

type MyInterface[T any, X II] interface {
	Get(ctx context.Context, name string) (T, error)
	Set(ctx context.Context, name string, value T) error
	Other(si SecondInterface) int
	Other2(ti ThirdInterface[T]) int
	Data()
	internal() bool
	CloseNotify() <-chan bool
	Unnamed(bool, string)
	XGet(ss *SI) *SI
}

type SecondInterface interface {
	Grab(ctx context.Context, name string) (any, error)
	Put(ctx context.Context, name string, value any) error
}

type ThirdInterface[H any] interface {
	Grab(ctx context.Context, name string) (any, error)
}
