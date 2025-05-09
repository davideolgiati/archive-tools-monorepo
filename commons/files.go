package commons

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type FileSize struct {
	Value int16
	Unit  string
}

type File struct {
	Name          string
	FormattedSize FileSize
	Hash          string
	Size          int64
}

func Compare_files(a *File, b *File) bool {
	return a.Hash < b.Hash && a.Size < b.Size
}

func Hash_file(filepath string, quick_flag bool) string {
	var err error
	var page_size int64 = int64(os.Getpagesize())

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
		if left_size > 1000000 {
			left_size = 1000000
		}
	} else {
		file_hash = sha1.New()
	}

	defer file_pointer.Close()

	r := bufio.NewReader(file_pointer)

	buf := make([]byte, page_size)
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
	return fmt.Sprintf("%x", sum)
}

func Get_human_reabable_size_async(size int64) FileSize {
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

func Current_user_has_read_right_on_file(obj *os.FileInfo) bool {
	read_bit_mask := fs.FileMode(0444)
	file_permission_bits := (*obj).Mode().Perm()

	return (file_permission_bits & read_bit_mask) == read_bit_mask
}

func Is_file_symbolic_link(path *string) bool {
	dst, err := filepath.EvalSymlinks(*path)

	if err != nil {
		return false
	}

	return *path != dst
}

func Is_file_a_device(obj *os.FileInfo) bool {
	return (*obj).Mode()&os.ModeDevice == os.ModeDevice
}

func Is_file_a_socket(obj *os.FileInfo) bool {
	return (*obj).Mode()&os.ModeSocket == os.ModeSocket
}

func Is_file_a_pipe(obj *os.FileInfo) bool {
	return (*obj).Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}
