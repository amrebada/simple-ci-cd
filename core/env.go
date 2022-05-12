package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func CopyDotEnvToDir(dir string, appId string) error {
	dotEnvPath := fmt.Sprintf(".%s.env", strings.ToLower(appId))
	info, err := os.Stat(dotEnvPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("dotenv is a directory")
	}
	if !info.Mode().IsRegular() {
		return errors.New("dotenv is not a regular file")
	}

	source, err := os.Open(dotEnvPath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(fmt.Sprintf("%s/.env", dir))
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
