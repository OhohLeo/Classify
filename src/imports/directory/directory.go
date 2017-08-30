package directory

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ohohleo/classify/data"
	"github.com/ohohleo/classify/imports"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Directory struct {
	Path        string `json:"path"`
	IsRecursive bool   `json:"is_recursive"`
	exiftoolCmd string
	isRunning   bool
	needToStop  bool
}

func ToBuild() imports.BuildImport {

	return imports.BuildImport{
		CheckConfig: func(config map[string][]string) (err error) {

			// For all specified directories
			for _, directories := range config {

				// All authorised path
				for _, path := range directories {

					// Check we have an existing directory
					if _, err = os.Stat(path); os.IsNotExist(err) {
						return
					}
				}
			}

			return
		},
		Create: func(input json.RawMessage, config map[string][]string, collections []string) (i imports.Import, params interface{}, err error) {
			var directory Directory
			err = json.Unmarshal(input, &directory)
			if err != nil {
				return
			}

			err = directory.Check(config, collections)
			if err != nil {
				return
			}

			i = &directory
			return
		},
	}
}

func (r *Directory) GetRef() imports.Ref {
	return imports.DIRECTORY
}

func (r *Directory) Check(config map[string][]string, collections []string) error {

	// Check we have an existing directory
	if _, err := os.Stat(r.Path); os.IsNotExist(err) {
		return err
	}

	// Check if exiftool exists
	cmd, err := exec.LookPath("exiftool")
	if err == nil {

		// Store exiftool command
		r.exiftoolCmd = cmd
	}

	// No config file : accept all
	if len(config) == 0 {
		return nil
	}

	// Check that the directory is in the global directories
	globalPaths, ok := config["*"]
	if ok {
		for _, path := range globalPaths {

			fmt.Printf("PATH %s => %s\n", r.Path, path)
			if r.Path == path {
				return nil
			}
		}
	}

	// Check that the directory is authorised for all specified collections
	for _, name := range collections {

		paths, ok := config[name]
		if ok == false {
			continue
		}

		for _, path := range paths {
			if r.Path == path {
				return nil
			}
		}
	}

	return errors.New("invalid or unauthorised import path '" + r.Path + "'")
}

// Return a channel of files in the directory
func (r *Directory) Start() (chan data.Data, error) {

	c := make(chan data.Data)

	// Check if the analysis is not already going on
	if r.isRunning {
		return c, fmt.Errorf("import 'directory' already started!")
	}

	// Analysis is starting
	r.isRunning = true

	// Reset stop process
	r.needToStop = false

	go func() {

		r.readDirectory(c, r.Path, r.IsRecursive)
		close(c)

		// Analysis is over
		r.isRunning = false
	}()

	return c, nil
}

func (r *Directory) Stop() error {
	r.needToStop = true
	return nil
}

func (r *Directory) readDirectory(c chan data.Data, path string, isRecursive bool) {

	// Read directory
	files, _ := ioutil.ReadDir(path)

	for _, f := range files {

		if r.needToStop {
			break
		}

		fullpath := path + "/" + f.Name()

		if f.IsDir() {

			// Read recursively
			if isRecursive {
				r.readDirectory(c, fullpath, isRecursive)
			}

			continue
		}

		fullname := f.Name()
		extension := filepath.Ext(fullname)

		file := &data.File{
			Name:      strings.TrimRight(fullname, extension),
			FullName:  fullname,
			Extension: extension,
			Path:      path,
			FullPath:  fullpath,
			FileInfo:  f,
		}

		// Search for file header infos
		if r.exiftoolCmd != "" {
			r.Analyse(r.exiftoolCmd, file)
		}

		// Send file info through channel
		c <- file
	}
}

func (r *Directory) Eq(new imports.Import) bool {
	newDirectory, _ := new.(*Directory)
	return r.Path == newDirectory.Path &&
		r.IsRecursive == newDirectory.IsRecursive
}

func (r *Directory) Analyse(cmdStr string, file *data.File) {

	fullpath := file.FullPath

	// Prepare command
	cmd := exec.Command(cmdStr, fullpath)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating StdoutPipe for '%s' [%s]: %s\n",
			cmdStr, fullpath, err.Error())
		return
	}

	// Analyse response
	scanner := bufio.NewScanner(cmdReader)
	go func() {

		if file.Infos == nil {
			file.Infos = make(map[string]string)
		}

		for scanner.Scan() {
			// Get result line by line
			res := strings.SplitN(scanner.Text(), ":", 2)
			key := strings.TrimSpace(res[0])
			value := strings.TrimSpace(res[1])

			// Store infos
			file.Infos[key] = value
		}
	}()

	// Execute the command
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error starting '%s' [%s]: %s\n", cmdStr, fullpath, err.Error())
		return
	}

	// Wait for the answer
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Error waiting '%s' [%s]: %s\n", cmdStr, fullpath, err.Error())
		return
	}

	// TODO : Check when file not recognized
}
