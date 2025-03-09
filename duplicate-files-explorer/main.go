package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type FileSize struct {
	value int16;
	unit string;
}

type File struct {
	name string;
	size FileSize;
	hash string;
}

func hash_file(basepath string, filename string, c chan string) {
	filepath := filepath.Join(basepath, filename)
	input, err := os.ReadFile(filepath)
	
	if err != nil {
		fmt.Print(err)
	}

	hash := md5.New()
	hash.Write(input)
	sum := hash.Sum(nil)

	c <- fmt.Sprintf("%x", sum)
}

func get_human_reabable_size(size int64, c chan FileSize) {
	file_size := size
	sizes_array := [4]string{"b", "Kb", "Mb", "Gb"}
	size_index := 0

	for size_index < 4 && file_size > 1000 {
		file_size /= 1000
		size_index++
	}

	output := FileSize{value: int16(file_size), unit: sizes_array[size_index]}

	c <- output	
}

func process_file_entry(basedir string, entry fs.FileInfo, c chan File) {
	hash_channel := make(chan string)
	size_channel := make(chan FileSize)
	go hash_file(basedir, entry.Name(), hash_channel)
	go get_human_reabable_size(entry.Size(), size_channel)

	file_size_info := <- size_channel 
	hash := <- hash_channel

	output := File{
		name: filepath.Join(basedir, entry.Name()),
		size: file_size_info,
		hash: hash,
	}

	c <- output

	/*
	c <- fmt.Sprintf(
		"file: %-55s %6s    %v\n", 
		filepath.Join(basedir, entry.Name()), 
		human_reabable_size, hash,
	)
	*/
}

func main() {
	var basedir string

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	entries, err := os.ReadDir(basedir)
    
    	if err != nil {
		log.Fatal(err)
    	}

    	channel := make(chan File)
    	counter := 0
 
    	for _, entry := range entries {
		entry_info, err := entry.Info()

		if err != nil {
			log.Fatal(err)
		}

		if(entry_info.Mode().IsRegular()) {
			go process_file_entry(basedir, entry_info, channel)
			counter++
		}
    	}

    	for i := 0; i < counter; i++ {
		data := <-channel
		fmt.Printf(
			"file: %-55s %6d %2s    %v\n",
			data.name, data.size.value, data.size.unit, data.hash,
		)
    	}
}