package main

import (
	"crypto/sha256"
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

	hash := sha256.New()
	hash.Write(input)
	sum := hash.Sum(nil)

	c <- fmt.Sprintf("%x", sum)
}

func process_file_entry(basedir string, entry os.DirEntry, c chan string) {
	if(!entry.Type().IsDir()) {
		info, err := entry.Info()

		if err != nil {
			log.Fatal(err)
		}

		hash_channel := make(chan string)
		go hash_file(basedir, info.Name(), hash_channel)

		hash := <- hash_channel

		c <- fmt.Sprintf(
			"file: %s/%-25s %6v Kb    %v\n", 
			basedir,
			info.Name(), 
			info.Size()/1000,
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