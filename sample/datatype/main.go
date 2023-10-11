package main

import (
	"errors"
	"fmt"

	"github.com/RangelReale/qdiimpl/sample/datatype/idata"
)

// -type=SampleData -data="*github.com/RangelReale/qdiimpl/sample/datatype/idata.IData" -overwrite=true

func main() {
	d := NewDebugSampleData(
		WithDebugSampleDataData(&idata.IData{Name: "xname", Value: "xvalue"}),
		WithDebugSampleDataGet(func(debugCtx *DebugSampleDataContext, name string) (any, error) {
			if name == debugCtx.Data.Name {
				return debugCtx.Data.Value, nil
			}
			return nil, errors.New("not found")
		}),
	)

	v, err := d.Get("xname")
	if err != nil {
		fmt.Printf("error: %s", err)
	} else {
		fmt.Println(v)
	}

	v, err = d.Get("x")
	if err != nil {
		fmt.Printf("error: %s", err)
	} else {
		fmt.Println(v)
	}
}
