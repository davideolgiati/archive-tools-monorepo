package commons

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

// func TestCompareFiles(t *testing.T) {
// 	t.Run("Template", func(t *testing.T) {
// 	})
// }

func new_file(data string) string {
	content := []byte(data)
	tmpfile, err := ioutil.TempFile("", "unit_test_file_")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile.Name()
}

func toOctal(number int) int {
	// defining variables and assigning them values
	octal := 0
	counter := 1
	remainder := 0
	for number != 0 {
	   remainder = number % 8
	   number = number / 8
	   octal += remainder * counter
	   counter *= 10
	}
	return octal
     }

func TestFileHash(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		test_data := "The quick brown fox jumps over the lazy dog"

		my_test_file := new_file(test_data)
		defer os.Remove(my_test_file)

		expected := "2fd4e1c67a2d28fced849ee1bb76e7391b93eb12"
		actual, err := Hash(my_test_file, int64(len(test_data)))

		if err != nil {
			panic(fmt.Sprintf("FileHashTestSuite - Happy Path : %v", err))
		}

		if actual != expected {
			panic(fmt.Sprintf("FileHashTestSuite - Happy Path : expected : %s, got : %s", expected, actual))
		}
	})

	t.Run("Empty", func(t *testing.T) {
		test_data := ""

		my_test_file := new_file(test_data)
		defer os.Remove(my_test_file)

		expected := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
		actual, err := Hash(my_test_file, int64(len(test_data)))

		if err != nil {
			panic(fmt.Sprintf("FileHashTestSuite - Happy Path : %v", err))
		}

		if actual != expected {
			panic(fmt.Sprintf("FileHashTestSuite - Happy Path : expected : %s, got : %s", expected, actual))
		}
	})

	t.Run("No File", func(t *testing.T) {
		test_data := ""

		my_test_file := "/tmp/pippo"

		actual, err := Hash(my_test_file, int64(len(test_data)))

		if err == nil {
			panic("FileHashTestSuite - Happy Path : expected populated error got nil")
		}

		if actual != "" {
			panic(fmt.Sprintf("FileHashTestSuite - Happy Path : expected : \"\", got : %s", actual))
		}
	})

	t.Run("Wrong Size", func(t *testing.T) {
		test_data := "The quick brown fox jumps over the lazy dog"

		my_test_file := new_file(test_data)
		defer os.Remove(my_test_file)

		expected := "2fd4e1c67a2d28fced849ee1bb76e7391b93eb12"
		actual, err := Hash(my_test_file, int64(len(test_data)*2))

		if err != nil {
			panic(fmt.Sprintf("FileHashTestSuite - Happy Path : %v", err))
		}

		if actual != expected {
			panic(fmt.Sprintf("FileHashTestSuite - Happy Path : expected : %s, got : %s", expected, actual))
		}
	})
}

func TestCheckReadRightsOnFile(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		valid_user_read_perm := []int{0400, 0500, 0600, 0700, 0000}
		valid_group_read_perm := []int{040, 050, 060, 070, 000}
		valid_others_read_perm := []int{04, 05, 06, 07}

		test_file := new_file("my file")

		for _, user_perm := range valid_user_read_perm {
			for _, group_perm := range valid_group_read_perm {
				for _, others_perm := range valid_others_read_perm {
					current_perm := user_perm + group_perm + others_perm
					os.Chmod(test_file, os.FileMode(current_perm))
					stats, err := os.Stat(test_file)

					if err != nil {
						panic(err)
					}

					if !Check_read_rights_on_file(&stats) {
						panic(fmt.Sprintf("Expected %d to be a valid read permission", current_perm))
					}
				}
			}
		}
	})

	t.Run("No Read Rights", func(t *testing.T) {
		valid_user_read_perm := []int{0100, 0200, 0300, 0000}
		valid_group_read_perm := []int{010, 020, 030, 000}
		valid_others_read_perm := []int{01, 02, 03}

		test_file := new_file("my file")

		for _, user_perm := range valid_user_read_perm {
			for _, group_perm := range valid_group_read_perm {
				for _, others_perm := range valid_others_read_perm {
					current_perm := user_perm + group_perm + others_perm
					os.Chmod(test_file, os.FileMode(current_perm))
					stats, err := os.Stat(test_file)

					if err != nil {
						panic(err)
					}

					if Check_read_rights_on_file(&stats) {
						panic(fmt.Sprintf("Expected %d not to be a valid read permission", current_perm))
					}
				}
			}
		}
	})
}
