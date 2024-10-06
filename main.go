package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	PackPath = `C:\texlive\2024\texmf-dist\tex\`
	FontPath = `C:\texlive\2024\texmf-dist\fonts\`
	DocPath  = `C:\texlive\2024\texmf-dist\doc\`
)

type PackFile struct {
	path    string
	author  string
	version string
}

func newpackfile(path, author, version string) PackFile {
	return PackFile{
		path:    path,
		author:  author,
		version: version,
	}
}

func (p PackFile) String() string {
	return fmt.Sprintf("path: %v\nauthor: %v\nversion: %v\n", p.path, p.author, p.version)
}

func init() {
	// check the texlive path from system environment
	envPath := os.Getenv("path")
	if exists := strings.Contains(envPath, "texlive"); !exists {
		fmt.Println("couldn't find texlive path in system env")
		os.Exit(2)
	}
}

func readPackInfo(pack string) (author, version string) {
	file, _ := os.ReadFile(pack)
	verRegex := regexp.MustCompile(`(?i)\\ProvidesPackage\{([^\}]+)\}\[([^\]]+)\]`)
	findver := verRegex.FindSubmatch(file)
	for len(findver) < 3 {
		return "", ""
	}
	version = string(findver[2])

	authRegex := regexp.MustCompile(`(?i)\%*\s*Copyright\s*\(C\)\s*[\d+\-\,\s]+\s([^\n]+)`)
	findAuthor := authRegex.FindAllSubmatch(file, -1)
	for _, v := range findAuthor {
		author = string(v[1])
	}
	return
}

func findPackage(packChan chan PackFile, pack string) {
	var packFile PackFile
	// just ignore the error message here because find also retrun info
	_ = filepath.Walk(PackPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.Contains(info.Name(), pack) {
			author, version := readPackInfo(path)
			packFile = newpackfile(path, author, version)
			return fmt.Errorf("package found %s", info.Name())
		}
		return nil
	})
	packChan <- packFile
}

func findDoc(docChan chan string, docName string) {
	docName = strings.Split(docName, ".")[0]
	_ = filepath.Walk(DocPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), docName) {
			docChan <- path
			return errors.New("search done")
		}
		return nil
	})
}

func findFont(fontName string) (fontPath string) {
	_ = filepath.Walk(FontPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.EqualFold(info.Name(), fontName) {
			fontPath = path
			return errors.New("search done")
		}
		return nil
	})
	return
}

func texPackage(packName string) {
	packChan := make(chan PackFile)
	docChan := make(chan string)

	go findPackage(packChan, packName)
	go findDoc(docChan, packName)

	select {
	case pack := <-packChan:
		fmt.Println(pack)
	case doc := <-docChan:
		fmt.Println(doc)
	}
}

func texFont(fontName string) {
	result := findFont(fontName)
	fmt.Println("path:", result)
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println("elasped time:", time.Since(start))
	}()

	fmt.Println("app for search TeX packages in local system")

	var packName string
	var fontName string

	flag.StringVar(&packName, "package", "test.sty", "provide a package name with ext to search in texlive dir.")
	flag.StringVar(&fontName, "font", "test.tfm", "provide the font name with ext to search")
	flag.Parse()

	if packName != "test.sty" {
		texPackage(packName)
	}
	if fontName != "test.tfm" {
		texFont(fontName)
	}
}
