package util

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"golang.org/x/sys/unix"
)

func GetTimestamp() (timestamp uint64) {
	return uint64(time.Now().Unix())
}

// Path and file functions
func GetHomeDir() (path string, err error) {
	user, err := user.Current()
	if err != nil {
		return path, err
	}
	return user.HomeDir, nil
}

func PathExistsAndIsWritable(path string) (err error) {
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("the path \"%s\" does not exist - please choose another path", path)
	}
	ok := unix.Access(path, unix.W_OK)
	if ok != nil {
		return fmt.Errorf("the path \"%s\" is not writable - please choose another path", path)
	}
	return nil
}

func FileExists(path string) (exists bool) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadFile(filename string) (string, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read \"%s\"", filename)
	}
	return string(bytes), nil
}

func DeleteFile(filename string) (err error) {
	if FileExists(filename) {
		err = os.Remove(filename)
		if err != nil {
			return fmt.Errorf("failed to remove the file \"%s\", %s", filename, err)
		}
	}
	return nil
}
