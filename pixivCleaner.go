package main

import (
	"fmt"
	"io"
	"io/ioutil"
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

// func treatFolder(subdir timeSlice) {
// 	dateSorted := make(timeSlice, 0, len(subdir))
// 	for _, d := range subdir {
// 		dateSorted = append(dateSorted, d)
// 	}
// 	sort.Sort(sort.Reverse(dateSorted))
// 	for _, v := range dateSorted {
// 		// fmt.Println(v.name, "->", dateSorted[0].name+"\\")
// 		content, _ := ioutil.ReadDir(v.name)
// 		for _, v2 := range content {
// 			if v2.IsDir() {
// 				continue
// 			}
// 			// fmt.Println(v.name+"\\"+v2.Name(), "->", dateSorted[0].name+"\\"+v2.Name())
// 			os.Rename(v.name+"\\"+v2.Name(), dateSorted[0].name+"\\"+v2.Name())
// 		}
// 	}
// }

func copyFile(source string, dest string) error {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()
	// _, err = os.Rename(destfile, sourcefile)
	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
			err = os.Chtimes(dest, sourceinfo.ModTime(), sourceinfo.ModTime())
		}
		sourcefile.Close()
		err = os.Remove(source)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func moveFile(source string, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		err := os.Remove(source)
		if err != nil {
			panic(err)
		}
	} else {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			panic(err)
		}
		err = os.Rename(source, dest)
		if err != nil {
			panic(err)
		}
		err = os.Chtimes(dest, sourceinfo.ModTime(), sourceinfo.ModTime())
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func copyDir(source string, dest string) error {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	// create dest dir
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
		fmt.Println(sourcefilepointer, "->", destinationfilepointer)
		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
			sourceinfo, err := os.Stat(source)
			if err != nil {
				err = os.Chmod(dest, sourceinfo.Mode())
				err = os.Chtimes(dest, sourceinfo.ModTime(), sourceinfo.ModTime())
			}
			fmt.Println("Removing folder: ", sourcefilepointer)
			err = os.Remove(sourcefilepointer)
			if err != nil {
				panic(err)
			}
		} else {
			// perform copy
			//err = copyFile(sourcefilepointer, destinationfilepointer)
			err = moveFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// func treatFolder(subdir timeSlice) error {
// 	dateSorted := make(timeSlice, 0, len(subdir))
// 	for _, d := range subdir {
// 		dateSorted = append(dateSorted, d)
// 	}
// 	sort.Sort(sort.Reverse(dateSorted))
// 	dest, err := os.Open(dateSorted[0].name)
// 	if err != nil {
// 		return err
// 	}
// 	defer dest.Close()
// 	for _, v := range dateSorted {
// 		if v.name == dateSorted[0].name {
// 			continue
// 		}
// 		fmt.Println("Copying folder: ", v.name)
// 		dir, err := os.Open(v.name)
// 		if err != nil {
// 			return err
// 		}
// 		content, err := dir.Readdir(-1)
// 		if err != nil {
// 			return err
// 		}
// 		for _, obj := range content {
// 			sourcePtr := v.name + "\\" + obj.Name()
// 			destPtr := dateSorted[0].name + "\\" + obj.Name()
// 			if obj.IsDir() {
// 				copyDir(sourcePtr, destPtr)
// 			} else {
// 				copyFile(sourcePtr, destPtr)
// 			}
// 		}
// 		fmt.Println("Removing folder: ", dir.Name())
// 		dir.Close()
// 		err = os.Remove(dir.Name())
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	return nil
// }

func main() {
	root, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	if !strings.Contains(root, root) {
		fmt.Println("Error: not a pixiv folder")
		os.Exit(1)
	}
	firstLevel, err := ioutil.ReadDir(root)
	if err != nil {
		panic(err)
	}
	for _, v1 := range firstLevel {
		_ = os.Chdir(root)
		if v1.IsDir() {
			err = os.Chdir(v1.Name())
			if err != nil {
				panic(err)
			}
			v1Path, err := filepath.Abs(".")
			if err != nil {
				panic(err)
			}
			secondLevel, err := ioutil.ReadDir(v1Path)
			if err != nil {
				panic(err)
			}
			subdir := make(timeSlice, 0)
			for _, v2 := range secondLevel {
				if v2.IsDir() {
					v2Path, err := filepath.Abs(v2.Name())
					if err != nil {
						panic(err)
					}
					subdir = append(subdir, subdirStruct{date: v2.ModTime(), name: v2Path})
				}
			}
			// err = treatFolder(subdir)
			dateSorted := make(timeSlice, 0, len(subdir))
			for _, d := range subdir {
				dateSorted = append(dateSorted, d)
			}
			sort.Sort(sort.Reverse(dateSorted))
			dest, err := os.Open(dateSorted[0].name)
			if err != nil {
				panic(err)
			}
			defer dest.Close()
			for _, v := range dateSorted {
				if v.name == dateSorted[0].name {
					continue
				}
				copyDir(v.name, dateSorted[0].name)
				fmt.Println("Removing folder: ", v.name)
				err = os.Remove(v.name)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
