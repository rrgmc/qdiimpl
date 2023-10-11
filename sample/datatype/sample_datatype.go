package main

type SampleData interface {
	Get(name string) (any, error)
}
