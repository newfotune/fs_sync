package cmd

import (
	"fmt"
	"fs_sync/models"
	"os/exec"
	"strings"
)

//TODO: Find a better way to check if file exist
func PathExist(userhost models.UserHost, path string) (bool, error) {
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", userhost.User, userhost.Host), "ls", path)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "No such file or directory") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func CreatePath(userhost models.UserHost, path string) error {
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", userhost.User, userhost.Host), "mkdir", "-p", path)

	_, err := cmd.CombinedOutput()
	return err
}

func GetEnv(userhost models.UserHost, envName string) (string, error) {
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", userhost.User, userhost.Host), "echo", "$"+envName)
	output, err := cmd.CombinedOutput()

	return strings.TrimSpace(string(output)), err
}
