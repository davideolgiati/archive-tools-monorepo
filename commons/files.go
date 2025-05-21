package commons

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
)

var sizes_array = [...]string{"b", "Kb", "Mb", "Gb"}
var page_size int64 = int64(os.Getpagesize())

type FileSize struct {
	Unit  string
	Value int16
}

type File struct {
	Name          string
	Hash          string
	FormattedSize FileSize
	Size          int64
}

func (file File) Format(f fmt.State, c rune) {
	f.Write([]byte(file.ToString()))
}

func (f File) ToString() string {
	return fmt.Sprintf(
		"%s %4d %2s %s",
		f.Hash,
		f.FormattedSize.Value,
		f.FormattedSize.Unit,
		f.Name,
	)
}

func Lower(a File, b File) bool {
	return a.Hash < b.Hash && a.Size < b.Size
}

func Equal(a File, b File) bool {
	return a.Hash == b.Hash && a.Size == b.Size
}

func Hash(filepath string, size int64) (string, error) {
	var err error = nil
	var hash_accumulator hash.Hash = sha1.New()
	var hash []byte

	file_pointer, err := os.Open(filepath)

	if err != nil {
		return "", err
	}

	defer file_pointer.Close()

	reader := bufio.NewReader(file_pointer)

	read_buffer := make([]byte, page_size)
	var read_size int = 0

	for size > 0 {
		read_size, err = reader.Read(read_buffer)

		if err != nil {
			if err == io.EOF {
				err = nil
			}
			size = 0
		}

		size = size - int64(read_size)
		hash_accumulator.Write(read_buffer[:read_size])
	}

	hash = hash_accumulator.Sum(nil)
	return fmt.Sprintf("%x", hash), err
}

func Format_file_size(size int64) FileSize {
	if size < 0 {
		panic("Format_file_size -- size is negative")
	}

	file_size := size
	size_index := 0

	for size_index < 3 && file_size > 1000 {
		file_size /= 1000
		size_index++
	}

	if file_size > 1000 && size_index != 3 {
		panic(fmt.Sprintf(
			"Format_file_size -- file_size is > 1000 and unit is  %s",
			sizes_array[size_index],
		))
	}

	output := FileSize{Value: int16(file_size), Unit: sizes_array[size_index]}

	return output
}

func Check_read_rights_on_file(obj *os.FileInfo) bool {
	if *obj == nil {
		panic("Check_read_rights_on_file -- obj is nil")
	}

	read_bit_mask := fs.FileMode(0444)
	file_permission_bits := (*obj).Mode().Perm()

	return (file_permission_bits & read_bit_mask) != fs.FileMode(0000)
}

func Is_symbolic_link(obj *os.FileInfo) bool {
	if *obj == nil {
		panic("Is_symbolic_link -- obj is empty")
	}

	return (*obj).Mode()&os.ModeSymlink != 0
}

func Is_a_device(obj *os.FileInfo) bool {
	if *obj == nil {
		panic("Is_a_device -- obj is nil")
	}

	return (*obj).Mode()&os.ModeDevice == os.ModeDevice
}

func Is_a_socket(obj *os.FileInfo) bool {
	if *obj == nil {
		panic("Is_a_socket -- obj is nil")
	}

	return (*obj).Mode()&os.ModeSocket == os.ModeSocket
}

func Is_a_pipe(obj *os.FileInfo) bool {
	if *obj == nil {
		panic("Is_a_pipe -- obj is nil")
	}

	return (*obj).Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}
