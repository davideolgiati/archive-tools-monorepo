package commons

import (
	"archive-tools-monorepo/dataStructures"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

var sizes_array = [...]string{"b", "Kb", "Mb", "Gb"}

type FileSize struct {
	Unit  *string
	Value int16
}

type File struct {
	Name                      string
	Hash                      dataStructures.Constant[string]
	FormattedataStructuresize FileSize
	Size                      int64
}

func (file File) Format(f fmt.State, c rune) {
	f.Write([]byte(file.ToString()))
}

func (f File) ToString() string {
	var b strings.Builder

	if f.Hash.Ptr() == nil {
		panic("Hash is a nil pointer")
	}

	b.WriteString(f.Hash.Value())
	for i := 0; i < 40-len(f.Hash.Value()); i++ {
		b.WriteByte(' ')
	}
	b.WriteByte(' ')

	// Right-align integer in 4-char space
	valStr := strconv.Itoa(int(f.FormattedataStructuresize.Value))
	for i := 0; i < 4-len(valStr); i++ {
		b.WriteByte(' ')
	}
	b.WriteString(valStr)

	b.WriteByte(' ')

	// Right-align unit in 2-char space
	unitStr := *f.FormattedataStructuresize.Unit
	for i := 0; i < 2-len(unitStr); i++ {
		b.WriteByte(' ')
	}
	b.WriteString(unitStr)

	b.WriteByte(' ')
	b.WriteString(f.Name)

	return b.String()
}

func SizeDescending(a File, b File) bool {
	if a.Size < 0 {
		panic("a.Size is negative")
	}

	if b.Size < 0 {
		panic("b.Size is negative")
	}

	if a.Size == b.Size {
		return true
	}

	return a.Size < b.Size
}

func HashDescending(a File, b File) bool {
	if a.Hash.Ptr() == nil {
		panic("a.Hash is a nil pointer")
	}

	if b.Hash.Ptr() == nil {
		panic("b.Hash is a nil pointer")
	}

	if a.Hash.Ptr() == b.Hash.Ptr() {
		return true
	}

	return a.Hash.Value() <= b.Hash.Value() || (a.Size < b.Size || a.Name < b.Name)
}

func Equal(a File, b File) bool {
	if a.Hash.Ptr() == nil {
		panic("a.Hash is a nil pointer")
	}

	if b.Hash.Ptr() == nil {
		panic("b.Hash is a nil pointer")
	}

	if a.Size < 0 {
		panic("a.Size is negative")
	}

	if b.Size < 0 {
		panic("b.Size is negative")
	}

	return a.Hash.Ptr() == b.Hash.Ptr() && a.Size == b.Size
}

func EqualByHash(a File, b File) bool {
	if a.Hash.Ptr() == nil {
		panic("a.Hash is a nil pointer")
	}

	if b.Hash.Ptr() == nil {
		panic("b.Hash is a nil pointer")
	}

	return a.Hash.Ptr() == b.Hash.Ptr()
}

func EqualBySize(a File, b File) bool {
	if a.Size < 0 {
		panic("a.Size is negative")
	}

	if b.Size < 0 {
		panic("b.Size is negative")
	}

	return a.Size == b.Size
}

func CalculateHash(filepath string) (string, error) {
	if filepath == "" {
		return "", fmt.Errorf("empty filepath")
	}

	file_pointer, err := os.Open(filepath)

	if err != nil {
		return "", err
	}

	if file_pointer == nil {
		return "", fmt.Errorf("file_pointer is nil")
	}

	defer file_pointer.Close()

	stats, err := file_pointer.Stat()

	if err != nil {
		return "", err
	}

	size := stats.Size()

	if size < 0 {
		return "", fmt.Errorf("size is not positive")
	}

	sha1h := sha1.New()
	_, err = io.Copy(sha1h, file_pointer)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sha1h.Sum(nil)), nil
}

func FormatFileSize(size int64) (FileSize, error) {
	if size < 0 {
		return FileSize{}, fmt.Errorf("size is negative")
	}

	file_size := float64(size)
	size_index := 0

	for size_index < 3 && file_size >= 1000 {
		file_size /= 1000
		size_index++
	}

	if file_size >= 1000 && size_index != 3 {
		return FileSize{}, fmt.Errorf(
			"file_size is > 1000 and unit is  %s",
			sizes_array[size_index],
		)
	}

	output := FileSize{Value: int16(file_size), Unit: &sizes_array[size_index]}

	return output, nil
}

func HasReadPermission(info *fs.FileInfo) bool {
	if info == nil {
		panic("obj is nil")
	}

	file_permission_bits := (*info).Mode().Perm()

	user_read_ok := file_permission_bits&fs.FileMode(0400) != fs.FileMode(0000)
	group_read_ok := file_permission_bits&fs.FileMode(0040) != fs.FileMode(0000)
	others_read_ok := file_permission_bits&fs.FileMode(0004) != fs.FileMode(0000)

	return user_read_ok || group_read_ok || others_read_ok
}

func IsSymbolicLink(info *fs.FileInfo) bool {
	if info == nil {
		panic("obj is nil")
	}

	return (*info).Mode()&os.ModeSymlink != 0
}

func IsDevice(info *fs.FileInfo) bool {
	if info == nil {
		panic("obj is nil")
	}

	return (*info).Mode()&os.ModeDevice == os.ModeDevice
}

func IsSocket(info *fs.FileInfo) bool {
	if info == nil {
		panic("obj is nil")
	}

	return (*info).Mode()&os.ModeSocket == os.ModeSocket
}

func IsPipe(info *fs.FileInfo) bool {
	if info == nil {
		panic("obj is nil")
	}

	return (*info).Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}
