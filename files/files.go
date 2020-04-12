package files

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	PROGRAM_HOME    = "/.fs_sync/"
	SYNC_LIST_FILE  = "/.fs_sync/.synclist"
	WHITE_LIST_FILE = "/.fs_sync/.whitelist"
)

var SUFFIX_WHITELIST = [...]string{".mod", ".sum"}

type FileSyncManager struct {
	SyncListFile        *os.File
	WhiteListFile       *os.File
	NumOfSyncFiles      int
	NumOfWhiteListFiles int
}

func init() {
	log.Println("Initializing files...")
	createFileIfNotExist()
}

func createFileIfNotExist() {
	err := os.MkdirAll(fmt.Sprintf("%s%s", os.Getenv("HOME"), PROGRAM_HOME), 0755)
	if err != nil {
		log.Fatal(err)
	}

	syncFile := fmt.Sprintf("%s%s", os.Getenv("HOME"), SYNC_LIST_FILE)
	whiteListFile := fmt.Sprintf("%s%s", os.Getenv("HOME"), WHITE_LIST_FILE)

	_, err = os.Stat(syncFile)
	if os.IsNotExist(err) {
		_, err = os.Create(syncFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = os.Stat(whiteListFile)
	if os.IsNotExist(err) {
		_, err = os.Create(whiteListFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func NewFileSyncManager() *FileSyncManager {
	fileSyncManager := new(FileSyncManager)
	fileSyncManager.SyncListFile = getFile(SYNC_LIST_FILE)
	fileSyncManager.WhiteListFile = getFile(WHITE_LIST_FILE)
	return fileSyncManager
}

func getFile(filePath string) *os.File {
	filePath = fmt.Sprintf("%s%s", os.Getenv("HOME"), filePath)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return file
}

func (this *FileSyncManager) GetFilesToSync() []string {
	lines := make([]string, 0)
	//TODO: why doesnt this work fith the file object?
	file, err := os.Open(this.SyncListFile.Name())
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func (this *FileSyncManager) AddFileToSynclist(filePath string) error {
	return addAllPathInDirToFile(filePath, this.SyncListFile)
}

func (this *FileSyncManager) AddFileToWhitelist(filePath string) error {
	return addAllPathInDirToFile(filePath, this.WhiteListFile)
}

func addAllPathInDirToFile(path string, file *os.File) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		for _, suffix := range SUFFIX_WHITELIST {
			if filepath.Ext(path) == suffix {
				return nil
			}
		}

		if isBinary(path) {
			return nil
		}

		log.Printf("Syncing file %s\n", path)
		_, err = file.WriteString(path + "\n")
		return err
	}

	log.Printf("Syncing dir %s\n", path)
	_, err = file.WriteString(path + "\n")

	filesInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	errorStrings := make([]string, 0)

	for _, fileInfo := range filesInfo {
		newPath := filepath.Join(path, fileInfo.Name())

		err = addAllPathInDirToFile(newPath, file)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	if len(errorStrings) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errorStrings, "\n"))
}

func (this *FileSyncManager) Close() error {
	errorStrings := make([]string, 0)

	err := this.SyncListFile.Close()
	errorStrings = append(errorStrings, err.Error())

	err = this.WhiteListFile.Close()
	errorStrings = append(errorStrings, err.Error())

	if len(errorStrings) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(errorStrings, "\n"))
}

func isBinary(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	bytes := make([]byte, 1024)
	n, err := file.Read(bytes)
	if err != nil {
		panic(err)
	}

	if n < 1024 {
		log.Printf("file %s has less than 1024 bytes", filePath)
	}

	var nullByte = []byte{0}

	for _, b := range bytes {
		if b == nullByte[0] {
			return true
		}
	}

	return false
}
