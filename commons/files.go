package commons

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

var sizes_array = [...]string{"b", "Kb", "Mb", "Gb"}
var page_size int64 = int64(os.Getpagesize())

type FileSize struct {
	Unit  *string
	Value int16
}

type File struct {
	Name          string
	Hash          *string
	FormattedSize FileSize
	Size          int64
}

func (file File) Format(f fmt.State, c rune) {
	f.Write([]byte(file.ToString()))
}

func (f File) ToString() string {
	var b strings.Builder
	b.WriteString(*f.Hash)
	b.WriteByte(' ')
	
	// Right-align integer in 4-char space
	valStr := strconv.Itoa(int(f.FormattedSize.Value))
	for i := 0; i < 4-len(valStr); i++ {
		b.WriteByte(' ')
	}
	b.WriteString(valStr)
	
	b.WriteByte(' ')
	
	// Right-align unit in 2-char space
	unitStr := *f.FormattedSize.Unit
	for i := 0; i < 2-len(unitStr); i++ {
		b.WriteByte(' ')
	}
	b.WriteString(unitStr)
	
	b.WriteByte(' ')
	b.WriteString(f.Name)
	
	return b.String()
}

func SizeDescending(a File, b File) bool {
	return a.Size < b.Size
}

func HashDescending(a File, b File) bool {
	return *a.Hash < *b.Hash
}

func Equal(a File, b File) bool {
	return a.Hash == b.Hash && a.Size == b.Size
}

func EqualByHash(a File, b File) bool {
	return *a.Hash == *b.Hash
}

func EqualBySize(a File, b File) bool {
	return a.Size == b.Size 
}

func Hash(filepath string, size int64) (string, error) {
	var err error = nil
	var hash_accumulator hash.Hash = sha1.New()
	var hash []byte

	file_pointer, err := os.Open(filepath)

	if err != nil {
		panic(err)
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
				size = 0
			} else {
				panic(err)
			}

		}

		size = size - int64(read_size)
		hash_accumulator.Write(read_buffer[:read_size])
	}

	hash = hash_accumulator.Sum(nil)
	return hex.EncodeToString(hash), err
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

	output := FileSize{Value: int16(file_size), Unit: &sizes_array[size_index]}

	return output
}

func Check_read_rights_on_file(obj *os.FileInfo) bool {
	if *obj == nil {
		panic("Check_read_rights_on_file -- obj is nil")
	}
	
	file_permission_bits := (*obj).Mode().Perm()

	user_read_ok := file_permission_bits & fs.FileMode(0400) != fs.FileMode(0000)
	group_read_ok := file_permission_bits & fs.FileMode(0040) != fs.FileMode(0000)
	others_read_ok := file_permission_bits & fs.FileMode(0004) != fs.FileMode(0000)

	return user_read_ok || group_read_ok || others_read_ok
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
