package commons

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func GetSHA1HashFromPath(filepath string) (string, error) {
	if filepath == "" {
		return "", fmt.Errorf("%w: empty filepath", os.ErrInvalid)
	}

	filePointer, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("error while generating hash: %w", err)
	}

	if filePointer == nil {
		return "", fmt.Errorf("%w: filePointer is nil", os.ErrInvalid)
	}

	defer func() {
		err = filePointer.Close()
		if err != nil {
			panic(err)
		}
	}()

	stats, err := filePointer.Stat()
	if err != nil {
		return "", fmt.Errorf("error while generating hash: %w", err)
	}

	size := stats.Size()

	if size < 0 {
		return "", fmt.Errorf("%w: file size is not positive", os.ErrInvalid)
	}

	sha1h := sha1.New()
	_, err = io.Copy(sha1h, filePointer)
	if err != nil {
		return "", fmt.Errorf("error while generating hash: %w", err)
	}

	return hex.EncodeToString(sha1h.Sum(nil)), nil
}
