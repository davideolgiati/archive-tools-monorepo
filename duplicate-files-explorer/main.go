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

func get_human_reabable_size(size int64, c chan string) {
	file_size := size
	sizes_array := [4]string{"b", "Kb", "Mb", "Gb"}
	size_index := 0

	for size_index < 4 && file_size > 1000 {
		file_size /= 1000
		size_index++
	}

	c <- fmt.Sprintf("%v %2s", file_size, sizes_array[size_index])	
}

func process_file_entry(basedir string, entry fs.FileInfo, c chan string) {
	hash_channel := make(chan string)
	size_channel := make(chan string)
	go hash_file(basedir, entry.Name(), hash_channel)
	go get_human_reabable_size(entry.Size(), size_channel)

	human_reabable_size := <- size_channel 
	hash := <- hash_channel

	c <- fmt.Sprintf(
		"file: %-55s %6s    %v\n", 
		filepath.Join(basedir, entry.Name()), 
		human_reabable_size, hash,
	)
}

func main() {
	var basedir string

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	entries, err := os.ReadDir(basedir)
    
    	if err != nil {
		log.Fatal(err)
    	}

    	channel := make(chan string)
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
		if(data != "") {
			fmt.Print(data)
		}
    	}
}