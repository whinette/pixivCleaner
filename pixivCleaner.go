package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type subdirStruct struct {
	date time.Time
	name string
}

type timeSlice []subdirStruct

func (s timeSlice) Less(i, j int) bool { return s[i].date.Before(s[j].date) }
func (s timeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s timeSlice) Len() int           { return len(s) }

// used to copy, choose to move files instead
// func copyFile(source string, dest string) error {
// 	sourcefile, err := os.Open(source)
// 	if err != nil {
// 		return err
// 	}
// 	destfile, err := os.Create(dest)
// 	if err != nil {
// 		return err
// 	}
// 	defer destfile.Close()
// 	// _, err = os.Rename(destfile, sourcefile)
// 	_, err = io.Copy(destfile, sourcefile)
// 	if err == nil {
// 		sourceinfo, err := os.Stat(source)
// 		if err != nil {
// 			err = os.Chmod(dest, sourceinfo.Mode())
// 			err = os.Chtimes(dest, sourceinfo.ModTime(), sourceinfo.ModTime())
// 		}
// 		sourcefile.Close()
// 		err = os.Remove(source)
// 		if err != nil {
// 			Errors.Panic(err)
// 		}
// 	}
// 	return nil
// }

func moveFile(source string, dest string, duplicate string) error {
	if _, err := os.Stat(dest); err == nil {
		// err := os.Remove(source)
		// if err != nil {
		// 	Errors.Panic(err)
		// }
		Info.Println(source, "-> Duplicate found ->", duplicate)
		sourceinfo, err := os.Stat(source)
		if err != nil {
			Errors.Panic(err)
		}
		err = os.Rename(source, duplicate)
		if err != nil {
			Errors.Panic(err)
		}
		err = os.Chmod(duplicate, sourceinfo.Mode())
		err = os.Chtimes(duplicate, sourceinfo.ModTime(), sourceinfo.ModTime())
		if err != nil {
			Errors.Panic(err)
		}
	} else {
		Info.Println(source, "->", dest)
		sourceinfo, err := os.Stat(source)
		if err != nil {
			Errors.Panic(err)
		}
		err = os.Rename(source, dest)
		if err != nil {
			Errors.Panic(err)
		}
		err = os.Chmod(dest, sourceinfo.Mode())
		err = os.Chtimes(dest, sourceinfo.ModTime(), sourceinfo.ModTime())
		if err != nil {
			Errors.Panic(err)
		}
	}
	return nil
}

func copyDir(source string, dest string, root string) error {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}
	directory, _ := os.Open(source)
	defer directory.Close()
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourcefilepointer := source + "\\" + obj.Name()
		destinationfilepointer := dest + "\\" + obj.Name()
		duplicatefilepointer := root + "\\.duplicates\\" + obj.Name()
		if obj.IsDir() {
			err = copyDir(sourcefilepointer, destinationfilepointer, root)
			if err != nil {
				Info.Println(err)
			}
			sourceinfo, err := os.Stat(source)
			if err != nil {
				err = os.Chmod(dest, sourceinfo.Mode())
				err = os.Chtimes(dest, sourceinfo.ModTime(), sourceinfo.ModTime())
			}
			Info.Println("Removing folder: ", sourcefilepointer)
			err = os.Remove(sourcefilepointer)
			if err != nil {
				Errors.Panic(err)
			}
		} else {
			err = moveFile(sourcefilepointer, destinationfilepointer, duplicatefilepointer)
			if err != nil {
				Info.Println(err)
			}
		}
	}
	return nil
}

var (
	Info   *log.Logger
	Errors *log.Logger
)

func initalise() (string, *os.File) {
	fd, err := os.OpenFile("pixivCleaner.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panic(err)
	}
	Info = log.New(io.MultiWriter(fd, os.Stdout), "INFO: ", log.Ldate|log.Ltime)
	Errors = log.New(io.MultiWriter(fd, os.Stderr), "ERROR: ", log.Ldate|log.Ltime)
	Info.Println("\n\n--- BEGINING NEW ENTRIES ---")
	root, err := filepath.Abs(".")
	if err != nil {
		Errors.Panic(err)
	}
	if !strings.Contains(root, root) {
		Info.Println("Error: not a pixiv folder")
		os.Exit(1)
	}
	if _, err := os.Stat(".duplicates"); os.IsNotExist(err) {
		rootinfo, _ := os.Stat(root)
		err = os.Mkdir(".duplicates", rootinfo.Mode())
		if err != nil {
			Errors.Panic(err)
		}
	}
	return root, fd
}

func main() {
	root, fd := initalise()
	defer fd.Close()
	firstLevel, err := ioutil.ReadDir(root)
	if err != nil {
		Errors.Panic(err)
	}
	for _, v1 := range firstLevel {
		_ = os.Chdir(root)
		if strings.HasPrefix(v1.Name(), ".") {
			Info.Printf("Folder %s is skipped\n", v1.Name())
			continue
		}
		if v1.IsDir() {
			err = os.Chdir(v1.Name())
			if err != nil {
				Errors.Panic(err)
			}
			v1Path, err := filepath.Abs(".")
			if err != nil {
				Errors.Panic(err)
			}
			secondLevel, err := ioutil.ReadDir(v1Path)
			if err != nil {
				Errors.Panic(err)
			}
			subdir := make(timeSlice, 0)
			for _, v2 := range secondLevel {
				if v2.IsDir() {
					v2Path, err := filepath.Abs(v2.Name())
					if err != nil {
						Errors.Panic(err)
					}
					subdir = append(subdir, subdirStruct{date: v2.ModTime(), name: v2Path})
				}
			}
			dateSorted := make(timeSlice, 0, len(subdir))
			for _, d := range subdir {
				dateSorted = append(dateSorted, d)
			}
			if dateSorted.Len() <= 0 {
				continue
			}
			sort.Sort(sort.Reverse(dateSorted))
			dest, err := os.Open(dateSorted[0].name)
			if err != nil {
				Errors.Panic(err)
			}
			defer dest.Close()
			for _, v := range dateSorted {
				if v.name == dateSorted[0].name {
					continue
				}
				copyDir(v.name, dateSorted[0].name, root)
				Info.Println("Removing folder: ", v.name)
				err = os.Remove(v.name)
				if err != nil {
					Errors.Panic(err)
				}
			}
		}
	}
}
