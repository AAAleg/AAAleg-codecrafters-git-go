package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
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
		blob, err := os.Open(path.Join(".git", "objects", blobSha[0:2], blobSha[2:]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading blob: %s\n", err)
			os.Exit(1)
		}
		defer func() {
			if err = blob.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing blob: %s\n", err)
				os.Exit(1)
			}
		}()

		dataReader, err := zlib.NewReader(blob)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error new zlib reader: %s\n", err)
			os.Exit(1)
		}
		defer func() {
			e := dataReader.Close()
			if err != nil && e != nil {
				fmt.Fprintf(os.Stderr, "close zlib reader: %s", e)
				os.Exit(1)
			}
		}()

		br := bufio.NewReader(dataReader)
		typ, err := br.ReadString(' ')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading blob data: %v", err)
			os.Exit(1)
		}

		typ = typ[:len(typ)-1] // cut ' '

		if typ != "blob" {
			fmt.Fprintf(os.Stderr, "Unsupported type: %v\n", typ)
			os.Exit(1)
		}

		sizeStr, err := br.ReadString('\000')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading data: %s\n", err)
			os.Exit(1)
		}

		sizeStr = sizeStr[:len(sizeStr)-1]
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parse size: %s\n", err)
			os.Exit(1)
		}

		_, err = io.CopyN(os.Stdout, br, size)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error copy data: %s\n", err)
			os.Exit(1)
		}
	case "hash-object":
		fileName := os.Args[3]

		typ := "blob"
		fileData, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
			os.Exit(1)
		}
		blobData := []byte{}
		blobData = append(blobData, []byte(typ)...)
		blobData = append(blobData, []byte(" ")...)
		blobData = append(blobData, []byte(strconv.Itoa(len(fileData)))...)
		blobData = append(blobData, []byte("\x00")...)
		blobData = append(blobData, fileData...)

		sha := sha1.New()
		sha.Write(blobData)
		hash := hex.EncodeToString(sha.Sum(blobData))
		from := len(hash) - 40
		hash = hash[from:]

		objectPath := path.Join(".git", "objects", string(hash[0:2]), string(hash[2:]))

		if err := os.MkdirAll(path.Dir(objectPath), os.ModeDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error mkdir file: %s\n", err)
			os.Exit(1)
		}

		blobFile, err := os.OpenFile(objectPath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating blob file: %s\n", err)
			os.Exit(1)
		}
		defer func() {
			if err = blobFile.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing blob: %s\n", err)
				os.Exit(1)
			}
		}()

		writer := zlib.NewWriter(blobFile)
		_, err = writer.Write(blobData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing blob file: %s\n", err)
			os.Exit(1)
		}
		defer func() {
			if err = writer.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing blob: %s\n", err)
				os.Exit(1)
			}
		}()
		fmt.Fprint(os.Stdout, "", string(hash))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
