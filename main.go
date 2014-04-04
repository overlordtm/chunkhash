package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileItem struct {
	info os.FileInfo
	path string
}

var (
	chunkSize = int64(4 * 1024 * 1024)
	h         hash.Hash
	fileCh    chan FileItem
	wg        sync.WaitGroup
)

func main() {

	root, _ := os.Getwd()
	hashAlg := "sha256"
	maxProc := 8

	flag.StringVar(&root, "root", root, "root dir to scan")
	flag.Int64Var(&chunkSize, "chunkSize", chunkSize, "chunk size in bytes")
	flag.StringVar(&hashAlg, "alg", hashAlg, "hash type")
	flag.IntVar(&maxProc, "maxProc", maxProc, "chansize")
	flag.Parse()

	fileCh = make(chan FileItem, maxProc)

	switch strings.ToLower(hashAlg) {
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	default:
		fmt.Println("Valid hashs: sha1, sha256, sha512, md5")
		return
	}

	go func() {
		filepath.Walk(root, walker)
		close(fileCh)
	}()

	for fi := range fileCh {
		wg.Add(1)
		go hashIt(fi.path, fi.info)
	}

	wg.Wait()
	return

}

func walker(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		fileCh <- FileItem{info, path}
	}

	return nil
}

func hashIt(path string, info os.FileInfo) {
	file, err := os.Open(path)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %s: %s\n", path, err)
	}

	for i := 0; i <= int(info.Size()/chunkSize); i = i + 1 {
		n, _ := io.CopyN(h, file, chunkSize)
		fmt.Printf("%s\t%d\t%d\t%x\n", path, info.Size(), n, h.Sum(nil))
		h.Reset()
	}
	wg.Done()
}
