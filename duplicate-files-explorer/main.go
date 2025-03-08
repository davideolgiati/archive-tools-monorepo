package main

import ("fmt"; "os"; "log")

func process_file_entry(entry os.DirEntry, c chan string) {
	if(entry.Type().IsDir()) {
		c <- fmt.Sprintf("%s: directory\n", entry.Name())
	} else {
		info, err := entry.Info()

		if err != nil {
			log.Fatal(err)
		}

		c <- fmt.Sprintf("%s: file - %v\n", info.Name(), entry.Type().Perm())
	}
}

func main() {
    entries, err := os.ReadDir("/")
    
    if err != nil {
        log.Fatal(err)
    }

    c := make(chan string)
    counter := 0
 
    for _, e := range entries {
            go process_file_entry(e, c)
	    counter++
    }
    
    for i := 0; i < counter; i++ {
	data := <-c
	fmt.Print(data)
    }
}