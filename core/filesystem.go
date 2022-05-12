package core

import "os"

func RemoveRepository(repositoryPath string) error {
	return os.RemoveAll(repositoryPath)
}
