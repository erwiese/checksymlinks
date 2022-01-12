package main

// see https://stackoverflow.com/questions/45022633/resolving-broken-symbolic-links

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	startTime := time.Now()

	fs := flag.NewFlagSet("checksymlinks", flag.ExitOnError)
	delBrokenLinks := fs.Bool("delete-broken", false, "If true, all broken symbolic links will be removed. Use with care! Defaults to false.")
	delAllLinks := fs.Bool("delete-all", false, "If true, all symbolic links will be removed. Use with care! Defaults to false.")

	fs.Usage = func() {
		fmt.Println(`checksymlinks - traverse a directory recursive and search for broken links.
	
Usage:
    checksymlinks [flags] <directory>
	
Flags:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
    Report broken links
    $ checksymlinks /home/user/xyz/dir1
	
    Delete broken links
    $ checksymlinks -delete-broken /home/user/xyz/dir1
	
Sources: https://github.com/erwiese/checksymlinks
Author: Erwin Wiesensarter`)
	}

	fs.Parse(os.Args[1:])
	argsNotParsed := fs.Args()
	if len(argsNotParsed) > 1 {
		fmt.Fprintf(os.Stderr, "unknown arguments: %s\n", strings.Join(argsNotParsed, " "))
		fs.Usage()
		os.Exit(1)
	} else if len(argsNotParsed) < 1 {
		fmt.Fprintf(os.Stderr, "No root path given\n")
		fs.Usage()
		os.Exit(1)
	}

	if *delBrokenLinks == true && *delAllLinks == true {
		fmt.Fprintf(os.Stderr, "Flags delBrokenLinks and delAllLinks are not allowed together\n")
		fs.Usage()
		os.Exit(1)
	}

	rootDir := argsNotParsed[0]
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		log.Fatalf("Path %s does not exist", rootDir)
	}

	err := os.Chdir(rootDir)
	if err != nil {
		log.Fatalf("Could not change to root-dir %s: %v", rootDir, err)
	}
	log.Printf("root dir: %s", rootDir)

	nofErrors := 0
	nofBrokenLinks := 0
	nofLinksRemoved := 0
	nofLinksInspected := 0

	// Traverse directory recursive, does not follow links
	// TODO use the new WalkDir function in Go1.16
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			log.Printf("visited dir: %q", path)
			return nil
		}

		//fmt.Printf("visited file or dir: %q\n", path)
		fi, err := os.Lstat(path)
		if err != nil {
			log.Fatalf("Could not get stat for %s: %v", path, err)
		}

		// If path is a symlink
		if fi.Mode()&os.ModeSymlink != 0 {
			nofLinksInspected++
			// remove link anyway
			if *delAllLinks == true {
				log.Printf("Remove link %s", path)
				err = os.Remove(path)
				if err != nil {
					nofErrors++
					log.Printf("Could not remove %s: %v", path, err)
				}
				nofLinksRemoved++
				return nil
			}

			// check if link is broken
			resolvedPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				log.Printf("broken link %s: %v", path, err)
				nofBrokenLinks++
				if *delBrokenLinks == true {
					log.Printf("Remove broken link %s", path)
					err = os.Remove(path)
					if err != nil {
						nofErrors++
						log.Printf("Could not remove broken link %s: %v", path, err)
					}
					nofLinksRemoved++
				}
			} else {
				log.Printf("symlink %s OK", resolvedPath)
			}

		}

		return nil
	})

	if err != nil {
		log.Fatalf("error walking the path %q: %v", rootDir, err)
	}

	// switch mode := fi.Mode(); {
	// case mode.IsRegular():
	// 	fmt.Println("regular file")
	// case mode.IsDir():
	// 	fmt.Println("directory")
	// case mode&os.ModeSymlink != 0:
	// 	fmt.Println("symbolic link")
	// case mode&os.ModeNamedPipe != 0:
	// 	fmt.Println("named pipe")
	// }

	log.Printf("%-16s %d", "inspected links:", nofLinksInspected)
	log.Printf("%-16s %d", "removed links:", nofLinksRemoved)
	log.Printf("%-16s %d", "broken links:", nofBrokenLinks)
	log.Printf("%-16s %d", "errors:", nofErrors)

	elapsed := time.Since(startTime)
	log.Printf("Execution time: %s", elapsed.String())

}
