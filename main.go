package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	if err := build(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context) error {
	// define build matrix
	oses := []string{"linux", "windows", "darwin", "android"}
	arches := []string{"amd64", "arm64"}

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// get reference to the local project
	src := client.Host().Directory(".")

	// create empty directory to put build outputs
	outputs := client.Directory()

	// get `golang` image
	golang := client.
		Container().
		From("golang:latest").
		WithDirectory("/src", src).
		WithWorkdir("/src")

	for _, goos := range oses {
		for _, goarch := range arches {
			path := fmt.Sprintf("build/%s/%s/", goos, goarch)

			build := golang.
				WithEnvVariable("GOOS", goos).
				WithEnvVariable("GOARCH", goarch).
				WithEnvVariable("CGO_ENABLED", "0")

			build = build.WithExec([]string{
				"go", "build",
				"-buildmode=pie", "-trimpath", "-ldflags", `-s -extldflags "-static"`,
				"-o", path,
			})

			outputs = outputs.WithDirectory(path, build.Directory(path))
		}
	}
	_, err = outputs.Export(ctx, ".")
	if err != nil {
		return err
	}

	return nil
}
