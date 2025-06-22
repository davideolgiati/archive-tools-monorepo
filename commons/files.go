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

var sizesArray = [...]string{"b", "Kb", "Mb", "Gb"}

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

func (file File) ToString() string {
	var b strings.Builder

	if file.Hash.Ptr() == nil {
		panic("Hash is a nil pointer")
	}

	b.WriteString(file.Hash.Value())
	for i := 0; i < 40-len(file.Hash.Value()); i++ {
		b.WriteByte(' ')
	}
	b.WriteByte(' ')

	// Right-align integer in 4-char space
	valStr := strconv.Itoa(int(file.FormattedataStructuresize.Value))
	for i := 0; i < 4-len(valStr); i++ {
		b.WriteByte(' ')
	}
	b.WriteString(valStr)

	b.WriteByte(' ')

	// Right-align unit in 2-char space
	unitStr := *file.FormattedataStructuresize.Unit
	for i := 0; i < 2-len(unitStr); i++ {
		b.WriteByte(' ')
	}
	b.WriteString(unitStr)

	b.WriteByte(' ')
	b.WriteString(file.Name)

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

func HashDescending(a *File, b *File) bool {
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

	filePointer, err := os.Open(filepath)

	if err != nil {
		return "", err
	}

	if filePointer == nil {
		return "", fmt.Errorf("filePointer is nil")
	}

	defer filePointer.Close()

	stats, err := filePointer.Stat()

	if err != nil {
		return "", err
	}

	size := stats.Size()

	if size < 0 {
		return "", fmt.Errorf("size is not positive")
	}

	sha1h := sha1.New()
	_, err = io.Copy(sha1h, filePointer)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sha1h.Sum(nil)), nil
}

func FormatFileSize(size int64) (FileSize, error) {
	if size < 0 {
		return FileSize{}, fmt.Errorf("size is negative")
	}

	fileSize := float64(size)
	sizeIndex := 0

	for sizeIndex < 3 && fileSize >= 1000 {
		fileSize /= 1000
		sizeIndex++
	}

	if fileSize >= 1000 && sizeIndex != 3 {
		return FileSize{}, fmt.Errorf(
			"fileSize is > 1000 and unit is  %s",
			sizesArray[sizeIndex],
		)
	}

	output := FileSize{Value: int16(fileSize), Unit: &sizesArray[sizeIndex]}

	return output, nil
}

func HasReadPermission(info *fs.FileInfo) bool {
	if info == nil {
		panic("obj is nil")
	}

	filePermissionsBits := (*info).Mode().Perm()

	userReadOk := filePermissionsBits&fs.FileMode(0400) != fs.FileMode(0000)
	groupReadOk := filePermissionsBits&fs.FileMode(0040) != fs.FileMode(0000)
	othersReadOk := filePermissionsBits&fs.FileMode(0004) != fs.FileMode(0000)

	return userReadOk || groupReadOk || othersReadOk
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
