package main

import (
	"context"
	"os"
)

func main() {
	ctx := context.Background()
	project := os.Getenv("PROJECT_ID")
	zone := os.Getenv("ZONE")
	instance := os.Getenv("INSTANCE")
	port := os.Getenv("PORT")
	orca, err := NewOrca(ctx, WithProject(project), WithZone(zone), WithPort(port), WithInstanceName(instance))
	if err != nil {
		panic(err)
	}
	orca.Run()
}
