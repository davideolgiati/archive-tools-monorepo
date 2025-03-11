package commons

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"bufio"
	"io"
)

type FileSize struct {
	Value int16;
	Unit string;
}

type File struct {
	Name string;
	Size FileSize;
	Hash string;
}

func Hash_file(basepath string, filename string, c chan string) {
	filepath := filepath.Join(basepath, filename)
	file_pointer, err := os.Open(filepath)
	hash := md5.New()
	
	if err != nil {
		fmt.Println("cannot able to read the file", err)
		return
	}
	     
	defer file_pointer.Close()

	r := bufio.NewReader(file_pointer)
	
	for {
		buf := make([]byte,4*1024) //the chunk size
		n, err := r.Read(buf) //loading chunk into buffer
		buf = buf[:n]

		if err != nil && err != io.EOF {
			panic(err)
		} else if err == io.EOF {
			break
		}
		
		hash.Write(buf)
	}

	sum := hash.Sum(nil)
	c <- fmt.Sprintf("%x", sum)
}

func Get_human_reabable_size(size int64) FileSize {
	file_size := size
	sizes_array := [4]string{"b", "Kb", "Mb", "Gb"}
	size_index := 0

	for size_index < 4 && file_size > 1000 {
		file_size /= 1000
		size_index++
	}

	output := FileSize{Value: int16(file_size), Unit: sizes_array[size_index]}

	return output	
}