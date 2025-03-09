package commons

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
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
	input, err := os.ReadFile(filepath)
	
	if err != nil {
		fmt.Print(err)
	}

	hash := md5.New()
	hash.Write(input)
	sum := hash.Sum(nil)

	c <- fmt.Sprintf("%x", sum)
}

func Get_human_reabable_size(size int64, c chan FileSize) {
	file_size := size
	sizes_array := [4]string{"b", "Kb", "Mb", "Gb"}
	size_index := 0

	for size_index < 4 && file_size > 1000 {
		file_size /= 1000
		size_index++
	}

	output := FileSize{Value: int16(file_size), Unit: sizes_array[size_index]}

	c <- output	
}