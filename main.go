package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 4 {
		log.Fatal("Usage: generator [source path] [implementation path] [mock implementation path]")
	}

	homeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	sourcePath := os.Args[1]
	implementPath := os.Args[2]
	mockPath := os.Args[3]

	ent := newEntitiesCollection(homeDir, sourcePath, implementPath, mockPath)
	ent.createFiles()
}
