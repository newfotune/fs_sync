package watcher

import (
	"fmt"
	"fs_sync/files"
	"fs_sync/models"
	"fs_sync/modules/cmd"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var userHost = models.UserHost{
	User: "vagrant",
	Host: "192.168.10.101",
}

type Watcher struct {
	FileSyncManager *files.FileSyncManager
	watcher         *fsnotify.Watcher
}

func New() *Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	return &Watcher{
		FileSyncManager: files.NewFileSyncManager(),
		watcher:         watcher,
	}
}

/**
Adds all the files in the directorty to the list of
 files to sync
*/
func (this *Watcher) AddDirs(fileDirs ...string) error {
	errorStrings := make([]string, 0)
	for _, file := range fileDirs {
		err := this.FileSyncManager.AddFileToSynclist(file)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
			continue
		}

		//TODO: investigate, this call. Watcher watches the directory
		//but it might not watch the content of the files which is what we
		//want to sync
		err = this.watcher.Add(file)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	if len(errorStrings) == 0 {
		return nil
	}

	return fmt.Errorf("%s", strings.Join(errorStrings, "\n"))
}

//TODO
func (this *Watcher) RemoveDir(dir string) {

}

func (this *Watcher) Start() {
	syncFiles := this.FileSyncManager.GetFilesToSync()
	cmdName := "rsync"
	for _, file := range syncFiles {
		remoteFile := file

		remoteHome, err := cmd.GetEnv(userHost, "HOME")
		if err != nil {
			log.Fatal(err)
		}

		//TODO: If not??
		if filepath.IsAbs(file) {
			remoteFile = strings.Replace(file, os.Getenv("HOME"), remoteHome, 1)
		}

		exist, err := cmd.PathExist(userHost, filepath.Dir(remoteFile))
		if err != nil {
			log.Fatalf("Error checking if path exists. %s", err)
		}

		if !exist {
			err := cmd.CreatePath(userHost, filepath.Dir(remoteFile))
			if err != nil {
				log.Fatalf("Error creating path in remote host. %s", err)
			}
		}
		this.watcher.Add(file)

		stat, err := os.Stat(file)
		if err != nil {
			panic(err)
		}

		if stat.IsDir() {
			continue
		}

		cmdArgs := strings.Split(
			fmt.Sprintf("-az --stats %s vagrant@192.168.10.101:%s", file, remoteFile), " ")

		log.Printf("syncing file %s", file)
		cmd := exec.Command(cmdName, cmdArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("error running rsync. %s\n%s", string(output), err)
		}
	}

	log.Printf("Starting watcherFunc...")
	go this.watcherFunc()
}

func (this *Watcher) watcherFunc() {
	for {
		select {
		case event, ok := <-this.watcher.Events:
			if !ok {
				log.Println("Not on in event")
				return
			}
			log.Println("event:", event)

			remoteHome, err := cmd.GetEnv(userHost, "HOME")
			if err != nil {
				log.Fatal(err)
			}

			cmdName := "rsync"
			if event.Op&fsnotify.Write == fsnotify.Write {
				f := strings.Replace(event.Name, os.Getenv("HOME"), remoteHome, 1)

				cmdArgs := strings.Split(
					fmt.Sprintf("-az --stats %s vagrant@192.168.10.101:%s", event.Name, f), " ")

				cmd := exec.Command(cmdName, cmdArgs...)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("error running rsync. %s\n%s", string(output), err)
				}
			}
		case err, ok := <-this.watcher.Errors:
			if !ok {
				log.Println("Not on in errors")
				return
			}
			log.Println("error:", err)
		}
	}
}

func (this *Watcher) Close() error {
	return this.watcher.Close()
}
