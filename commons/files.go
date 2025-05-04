package commons

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"hash"
	"hash/crc32"
)

type FileSize struct {
	Value int16
	Unit  string
}

type File struct {
	Name          string
	FormattedSize FileSize
	Hash          string
	Size 	      int64
}

func Hash_file(filepath string, quick_flag bool, c chan string) {
	var err error
	
	file_pointer, err := os.Open(filepath)
	var file_hash hash.Hash

	if err != nil {
		panic(err)
	}

	file_info, err := file_pointer.Stat()

	if err != nil {
		panic(err)
	}

	left_size := file_info.Size()

	if quick_flag {
		file_hash = crc32.New(crc32.IEEETable)
		if left_size > 4000 {
			left_size = 4000
		}
	} else {
		file_hash = md5.New()
	}

	defer file_pointer.Close()

	r := bufio.NewReader(file_pointer)

	buf := make([]byte, 2000)
	var read_size int

	for left_size > 0 {
		read_size, err = r.Read(buf)
		buf = buf[:read_size]

		left_size = left_size - int64(read_size)

		if err != nil && err != io.EOF {
			panic(fmt.Sprintf("%s\n\n", err))
		} else if err == io.EOF && left_size > 0 {
			left_size = 0
			//panic(fmt.Sprintf("left size is positive: %d", left_size))
		}

		file_hash.Write(buf)
	}

	sum := file_hash.Sum(nil)
	c <- fmt.Sprintf("%x", sum)
}

func Get_human_reabable_size_async(size int64, file_size_channel chan FileSize) {
	file_size := size
	sizes_array := [4]string{"b", "Kb", "Mb", "Gb"}
	size_index := 0

	for size_index < 4 && file_size > 1000 {
		file_size /= 1000
		size_index++
	}

	output := FileSize{Value: int16(file_size), Unit: sizes_array[size_index]}

	file_size_channel <- output
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
		return false
	}

	read_bit_mask := fs.FileMode(0444)
	file_permission_bits := info.Mode().Perm()

	return (file_permission_bits & read_bit_mask) == read_bit_mask
}

func Is_file_symbolic_link(fullpath *string) bool {
	info, err := os.Lstat(*fullpath)

	if err != nil {
		return false
	}

	return info.Mode()&os.ModeSymlink == os.ModeSymlink
}

func Is_file_a_device(fullpath *string) bool {
	info, err := os.Stat(*fullpath)

	if err != nil {
		return false
	}

	return info.Mode()&os.ModeDevice == os.ModeDevice
}

func Is_file_a_socket(fullpath *string) bool {
	info, err := os.Stat(*fullpath)

	if err != nil {
		return false
	}

	return info.Mode()&os.ModeSocket == os.ModeSocket
}

func Is_file_a_pipe(fullpath *string) bool {
	info, err := os.Stat(*fullpath)

	if err != nil {
		return false
	}

	return info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}
