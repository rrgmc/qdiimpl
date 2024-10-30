package main

import (
	"errors"
	"fmt"

	"github.com/rrgmc/qdiimpl/sample/datatype/idata"
)

// -type=SampleData -data-type="*github.com/rrgmc/qdiimpl/sample/datatype/idata.IData" -export-type=true -overwrite=true -name-prefix=QD -option-name-prefix=QD

func main() {
	d := NewQDSampleData(
		WithQDData(&idata.IData{Name: "xname", Value: "xvalue"}),
		WithQDGet(func(qdCtx *QDSampleDataContext, name string) (any, error) {
			qdCtx.NotSupported() // will ignore return values and call the next handler
			return nil, nil
		}),
		WithQDGet(func(qdCtx *QDSampleDataContext, name string) (any, error) {
			if name == qdCtx.Data.Name {
				return qdCtx.Data.Value, nil
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
