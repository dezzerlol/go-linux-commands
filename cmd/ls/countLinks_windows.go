//go:build windows
// +build windows

package ls

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
)

// syscall.Stat_t linux = syscall.Win32FileAttributeData on windows
// fix for l8r
func countLinks(dirPath string) (int, error) {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	list, err := f.Readdirnames(-1)
	f.Close()

	return len(list), err
}

func getFileOwners(info fs.FileInfo) (*user.User, *user.Group, error) {
	var UID int
	var GID int

	UID = os.Getuid()
	GID = os.Getgid()

	group, err := user.LookupGroupId(fmt.Sprintf("%d", GID))

	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	owner, err := user.LookupId(fmt.Sprintf("%d", UID))

	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	return owner, group, nil
}
