package main

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/master\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		blobSha := os.Args[3]
		blob, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", blobSha[0:2], blobSha[2:]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading blob: %s\n", err)
		}
		defer func() {
			if err = blob.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing blob: %s\n", err)
			}
		}()

		blobData := bufio.NewReader(blob)

		dataReader, err := zlib.NewReader(blobData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading blob data: %s\n", err)
		}
		defer dataReader.Close()

		objectType := make([]byte, 1)
		for {
			n, err := dataReader.Read(objectType)
			if n > 0 {
				t := string(objectType[:n])
				if t == "\x00" {
					break
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading type: %s\n", err)
				break
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading blob type: %s\n", err)
		}

		io.Copy(os.Stdout, dataReader)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
