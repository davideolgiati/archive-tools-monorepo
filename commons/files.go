package commons

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	datastructures "archive-tools-monorepo/dataStructures"
)

var sizesArray = [...]string{"b", "Kb", "Mb", "Gb"}

type FileSize struct {
	Unit  *string
	Value int16
}

type File struct {
	Hash datastructures.Constant[string]
	Name string
	Size int64
}

func (file *File) Format(f fmt.State, _ rune) {
	str, err := file.ToString()
	if err != nil {
		panic(err)
	}

	data := []byte(str)
	byteCount, err := f.Write(data)
	if err != nil {
		panic(err)
	}

	if byteCount != len(data) {
		panic("written data mismatch input data length")
	}
}

func (file *File) ToString() (string, error) {
	var b strings.Builder

	if file.Hash.Ptr() == nil {
		return "", fmt.Errorf("%w: Hash is a nil pointer", os.ErrInvalid)
	}

	b.WriteString(file.Hash.Value())
	for range 40 - len(file.Hash.Value()) {
		b.WriteByte(' ')
	}
	b.WriteByte(' ')

	// Right-align integer in 4-char space
	formattedFileSize, err := FormatFileSize(file.Size)
	if err != nil {
		return "", err
	}

	valStr := strconv.Itoa(int(formattedFileSize.Value))
	for range 4 - len(valStr) {
		b.WriteByte(' ')
	}
	b.WriteString(valStr)

	b.WriteByte(' ')

	// Right-align unit in 2-char space
	for range 2 - len(*formattedFileSize.Unit) {
		b.WriteByte(' ')
	}
	b.WriteString(*formattedFileSize.Unit)

	b.WriteByte(' ')
	b.WriteString(file.Name)

	return b.String(), nil
}

func (file *File) HashDescending(b *File) bool {
	if file.Hash.Ptr() == nil || b.Hash.Ptr() == nil {
		return false
	}

	if file.Hash.Ptr() == b.Hash.Ptr() {
		return true
	}

	return file.Hash.Value() <= b.Hash.Value() || (file.Size < b.Size || file.Name < b.Name)
}

func (file *File) Equal(other *File) bool {
	if file.Hash.Ptr() == nil || other.Hash.Ptr() == nil || file.Size < 0 || other.Size < 0 {
		return false
	}

	return file.Hash.Ptr() == other.Hash.Ptr() && file.Size == other.Size
}

func (file *File) EqualByHash(other *File) bool {
	if file.Hash.Ptr() == nil || other.Hash.Ptr() == nil {
		return false
	}

	return file.Hash.Ptr() == other.Hash.Ptr()
}

func StrongFileCompare(f1, f2 *File) bool {
	return f1.HashDescending(f2)
}

func WeakFileEqulity(f1, f2 *File) bool {
	return f1.EqualByHash(f2)
}

func StrongFileEquality(f1, f2 *File) bool {
	return f1.Equal(f2)
}

func FormatFileSize(size int64) (FileSize, error) {
	if size < 0 {
		return FileSize{}, fmt.Errorf("%w: size is negative", os.ErrInvalid)
	}

	fileSize := float64(size)
	sizeIndex := 0

	for sizeIndex < 3 && fileSize >= 1000 {
		fileSize /= 1000
		sizeIndex++
	}

	if fileSize >= 1000 && sizeIndex != 3 {
		return FileSize{}, fmt.Errorf(
			"%w: fileSize is > 1000 and unit is  %s",
			os.ErrInvalid, sizesArray[sizeIndex],
		)
	}

	output := FileSize{Value: int16(fileSize), Unit: &sizesArray[sizeIndex]}

	return output, nil
}
