package ls

import (
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/dezzerlol/go-linux-commands/internal/color"
	"github.com/spf13/cobra"
)

var (
	// Do not ignore entries starting with .
	all bool
	// Use long listing format
	long bool
	// File sizes in readable readable format
	readable bool
)

var (
	dirKey = "d"
	libKey = "l"
)

type File struct {
	ftype            string
	permission       string
	links            int
	user             string
	group            string
	size             int64
	modificationDate string
	name             string
}

func init() {
	Cmd.Flags().BoolVarP(&long, "long", "l", false, "Use long listing format")
	Cmd.Flags().BoolVarP(&all, "all", "a", false, "Do not ignore entries starting with .")
	Cmd.Flags().BoolVarP(&readable, "readable", "r", false, "Human readable file size")
}

// BUG: panic when trying to read files like swapfile.sys
var Cmd = &cobra.Command{
	Use:     "ls [OPTIONS] [FILEs]",
	Short:   "ls lists all files in directory",
	Long:    `ls lists all files in directory`,
	Version: "1.0.0",
	Args:    cobra.RangeArgs(0, 99),
	Run: func(cmd *cobra.Command, args []string) {
		path := "./"

		if len(args) > 0 {
			path = args[0]
		}

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered. Error:\n", r)
			}
		}()

		files, err := listFiles(path)

		if err != nil {
			log.Fatal(err)
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)

		for _, f := range files {
			output := fmt.Sprintf("%s\t", f.name)
			var fileSize interface{}

			if readable {
				fileSize = prettyByteSize(int(f.size))
			} else {
				fileSize = f.size
			}

			if long {
				output = fmt.Sprintf("%s\t%s\t%v\t%s\t%s\t%v\t%s\t%s",
					f.ftype,
					f.permission,
					f.links,
					f.group,
					f.user,
					fileSize,
					f.modificationDate,
					output)
			}

			fmt.Fprintln(writer, output)
		}

		fmt.Fprint(writer, "\n")
		writer.Flush()
	},
}

func listFiles(path string) ([]File, error) {
	var result []File

	files, err := os.ReadDir(path)

	if err != nil {
		log.Println(err)
		return []File{}, err
	}

	for _, d := range files {
		info, _ := d.Info()

		linksCount := 1
		ftype := strings.ToLower(string(d.Type().String()[0]))
		lastModification := info.ModTime().Format("Jan 2 15:04")
		permission := info.Mode().Perm().String()
		fileName := d.Name()

		if fileName[0] == '.' && !all {
			continue
		}

		if ftype == dirKey {
			fileName = color.ColorStr(fileName, color.Blue)
		}

		if ftype == libKey {
			fileName = color.ColorStr(fileName, color.Cyan)
		}

		if d.Type().IsDir() {
			links, err := countLinks(filepath.Join(path, d.Name()))

			if err != nil {
				log.Println(err)
				continue
			}

			linksCount = links
		}

		owner, group, err := getFileOwners(info)

		if err != nil {
			log.Println(err)
			continue
		}

		result = append(result, File{
			ftype:            ftype,
			permission:       permission,
			links:            linksCount,
			group:            group.Name,
			user:             owner.Username,
			size:             info.Size(),
			modificationDate: lastModification,
			name:             fileName,
		})
	}

	return result, err
}

func getFileOwners(info fs.FileInfo) (*user.User, *user.Group, error) {
	var UID int
	var GID int

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	} else {
		// we are not in linux, this won't work anyway in windows,
		// but maybe you want to log warnings
		UID = os.Getuid()
		GID = os.Getgid()
	}

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

func prettyByteSize(b int) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}
