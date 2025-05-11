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

var sizes_array = [...]string{"b", "Kb", "Mb", "Gb"}
var page_size int64 = int64(os.Getpagesize())

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

func (file *File) Format(f fmt.State, c rune) {
	f.Write([]byte(file.ToString()))
}

func (f *File) ToString() string {
	return fmt.Sprintf(
		"%s %4d %2s %s",
		f.Hash,
		f.FormattedSize.Value,
		f.FormattedSize.Unit,
		f.Name,
	)
}

func Lower(a *File, b *File) bool {
	return a.Hash < b.Hash && a.Size < b.Size
}

func Equal(a *File, b *File) bool {
	return a.Hash == b.Hash && a.Size == b.Size
}

func Hash(filepath *string, size int64, quick_flag bool) (string, error) {
	var err error = nil
	var hash_accumulator hash.Hash = crc32.New(crc32.IEEETable)
	var hash []byte

	file_pointer, err := os.Open(*filepath)

	if err != nil {
		return "", err
	}

	if quick_flag {
		size = page_size * 5
	} else {
		hash_accumulator = sha1.New()
	}

	defer file_pointer.Close()

	reader := bufio.NewReader(file_pointer)

	read_buffer := make([]byte, page_size)
	var read_size int

	for size > 0 {
		read_size, err = reader.Read(read_buffer)
		
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			size = 0
		}
		
		size = size - int64(read_size)
		read_buffer = read_buffer[:read_size]
		hash_accumulator.Write(read_buffer)
	}

	hash = hash_accumulator.Sum(nil)
	return fmt.Sprintf("%x", hash), err
}

func Format_file_size(size int64) FileSize {
	file_size := size
	size_index := 0

	for size_index < 3 && file_size > 1000 {
		file_size /= 1000
		size_index++
	}

	output := FileSize{Value: int16(file_size), Unit: sizes_array[size_index]}

	return output
}

func Check_read_rights_on_file(obj *os.FileInfo) bool {
	read_bit_mask := fs.FileMode(0444)
	file_permission_bits := (*obj).Mode().Perm()

	return (file_permission_bits & read_bit_mask) != fs.FileMode(0000)
}

func Is_symbolic_link(path *string) bool {
	dst, err := filepath.EvalSymlinks(*path)

	if err != nil {
		return false
	}

	return *path != dst
}

func Is_a_device(obj *os.FileInfo) bool {
	return (*obj).Mode()&os.ModeDevice == os.ModeDevice
}

func Is_a_socket(obj *os.FileInfo) bool {
	return (*obj).Mode()&os.ModeSocket == os.ModeSocket
}

func Is_a_pipe(obj *os.FileInfo) bool {
	return (*obj).Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}
