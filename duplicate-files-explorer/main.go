package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
)

func hash_file(basepath string, filename string, c chan string) {
	filepath := fmt.Sprintf("%s/%s", basepath, filename)
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

func process_file_entry(basedir string, entry os.DirEntry, c chan string) {
	if(!entry.Type().IsDir()) {
		info, err := entry.Info()

		if err != nil {
			log.Fatal(err)
		}

		hash_channel := make(chan string)
		size_channel := make(chan string)
		go hash_file(basedir, info.Name(), hash_channel)
		go get_human_reabable_size(info.Size(), size_channel)

		human_reabable_size := <- size_channel 
		hash := <- hash_channel

		c <- fmt.Sprintf(
			"file: %s/%-25s %6s    %v\n", 
			basedir, info.Name(), 
			human_reabable_size,
			hash,
		)
	} else {
		c <- ""
	}
}

func main() {
	basedir := "/home/davide"
	entries, err := os.ReadDir(basedir)
    
    	if err != nil {
		log.Fatal(err)
    	}

    	channel := make(chan string)
    	counter := 0
 
    	for _, entry := range entries {
		go process_file_entry(basedir, entry, channel)
		counter++
    	}
    
    	for i := 0; i < counter; i++ {
		data := <-channel
		if(data != "") {
			fmt.Print(data)
		}
    	}
}