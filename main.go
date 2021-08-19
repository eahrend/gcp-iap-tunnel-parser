package main

import (
	"context"
)

func main() {
	ctx := context.Background()
	orca, err := NewOrca(ctx)
	if err != nil {
		panic(err)
	}
	panic(orca.Run())
}
