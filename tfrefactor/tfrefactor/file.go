package tfrefactor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
)

// Note: This file and its methods are taken from https://github.com/minamijoyo/tfupdate/blob/master/tfupdate/file.go
// with the exception that they are renamed with the "Migrate" prefix

// MigrateFile migrates resources in a new single file.
// Optionally will generate a resulting CSV with the new resources and their parent.
// We use an afero filesystem here for testing.
func MigrateFile(fs afero.Fs, filename string, o Option) error {
	log.Printf("[DEBUG] check file: %s", filename)
	r, err := fs.Open(filename)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to open file: %s", err)
	}
	defer r.Close()

	w := &bytes.Buffer{}
	newResourceNames, err := MigrateHCL(r, w, filename, o)
	if err != nil {
		return err
	}

	// Write contents to destination file if migrations occurred.
	if len(newResourceNames) == 0 {
		log.Printf("[DEBUG] no migration file to create for %s", filename)
		return nil
	}

	outputFilename := strings.Replace(filename, ".tf", "_migrated.tf", 1)
	log.Printf("[INFO] new file: %s", outputFilename)
	migrated := w.Bytes()
	// We should be able to choose whether to format output or not.
	// However, the current implementation of (*hclwrite.Body).SetAttributeValue()
	// does not seem to preserve an original SpaceBefore value of attribute.
	// So, we need to format output here.
	result := hclwrite.Format(migrated)
	if err = afero.WriteFile(fs, outputFilename, result, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	// Write migrations to csv file
	if o.Csv {
		newFile, err := os.Create(strings.Replace(filename, ".tf", "_new_resources.csv", 1))
		log.Printf("[INFO] new file: %s", newFile.Name())
		if err != nil {
			return fmt.Errorf("[ERROR] error creating (%s): %s", newFile.Name(), err)
		}

		defer newFile.Close()

		for _, r := range newResourceNames {
			if _, err := newFile.WriteString(fmt.Sprintf("%s\n", r)); err != nil {
				log.Printf("[ERROR] error writing (%s) to file (%s): %s", r, newFile.Name(), err)
			}
		}
	}

	return nil
}

// MigrateDir migrates resources for files in a given directory.
// If a recursive flag is true, it checks and migrates recursively.
// skip hidden directories such as .terraform or .git.
// It also skips a file without .tf extension.
func MigrateDir(fs afero.Fs, dirname string, o Option) error {
	log.Printf("[DEBUG] check dir: %s", dirname)
	dir, err := afero.ReadDir(fs, dirname)
	if err != nil {
		return fmt.Errorf("failed to open dir: %s", err)
	}

	for _, entry := range dir {
		path := filepath.Join(dirname, entry.Name())

		// if a path of entry matches ignorePaths, skip it.
		if o.MatchIgnorePaths(path) {
			log.Printf("[DEBUG] ignore: %s", path)
			continue
		}

		if entry.IsDir() {
			// if an entry is a directory
			if !o.Recursive {
				// skip directory if a recursive flag is false
				continue
			}
			if strings.HasPrefix(entry.Name(), ".") {
				// skip hidden directories such as .terraform or .git
				continue
			}

			err := MigrateDir(fs, path, o)
			if err != nil {
				return err
			}

			continue
		}

		// if an entry is a file
		if filepath.Ext(entry.Name()) != ".tf" {
			// skip a file without .tf extension.
			continue
		}

		err := MigrateFile(fs, path, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// MigrateFileOrDir updates version constraints in a given file or directory.
func MigrateFileOrDir(fs afero.Fs, path string, o Option) error {
	isDir, err := afero.IsDir(fs, path)
	if err != nil {
		return fmt.Errorf("failed to open path: %s", err)
	}

	if isDir {
		// if an entry is a directory
		return MigrateDir(fs, path, o)
	}

	// if an entry is a file
	return MigrateFile(fs, path, o)
}
