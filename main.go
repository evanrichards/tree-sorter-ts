package main

import "github.com/evanrichards/tree-sorter-ts/internal/app"

// This file exists for backward compatibility with:
// go run github.com/evanrichards/tree-sorter-ts@latest
//
// The actual implementation is in internal/app/
// The canonical entrypoint is cmd/tree-sorter-ts/

func main() {
	app.Run()
}