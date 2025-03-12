package commons

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"bufio"
	"io"
	"io/fs"
	"errors"
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
		panic(err)
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

func Current_user_has_read_right_on_file(fullpath *string) bool {
	info, err := os.Stat(*fullpath)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		panic(err)
	}

	read_bit_mask := fs.FileMode(0444)
	file_permission_bits := info.Mode().Perm()
	
	return (file_permission_bits & read_bit_mask) == read_bit_mask
}