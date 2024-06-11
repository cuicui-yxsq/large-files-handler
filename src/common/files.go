package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Split files into chunks of the specified size.
func SplitFile(filePath string, chunkSize uint, outDir string) (outFilePaths []string, err error) {
	if chunkSize == 0 {
		err = fmt.Errorf("chunk size must be greater than 0")
		return
	}

	// open input file
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// ensure that the input file is a regular file
	stat, err := file.Stat()
	if err != nil {
		return
	}
	if !stat.Mode().IsRegular() {
		err = fmt.Errorf("not a regular file: %s", filePath)
		return
	}

	// if `outDir` is not specified, use current working directory
	if outDir == "" {
		outDir = "."
	} else {
		// otherwise, ensure that the output directory exists
		// if it does not exist, create it
		err = os.MkdirAll(outDir, 0755)
		if err != nil {
			return
		}
	}

	// split the input file
	basename := filepath.Base(filePath)
	buff := make([]byte, chunkSize)
	for i, eof := uint(0), false; !eof && (err == nil); i++ {
		// use an inner function so that
		// each output file is closed after writing to it has been completed
		err = func() (err error) {
			// read input file
			n, err := file.Read(buff)
			if err != nil {
				if err == io.EOF {
					eof = true
				} else {
					return
				}
			}
			if n == 0 {
				return // do not write empty files
			}

			// create output file
			outFilePath := filepath.Join(outDir, fmt.Sprintf("%s.%s%d", basename, SplitSuffix, i))
			outFile, err := os.Create(outFilePath)
			if err != nil {
				return
			}
			defer outFile.Close()

			// write output file
			_, err = outFile.Write(buff[:n])
			if err != nil {
				return
			}

			outFilePaths = append(outFilePaths, outFilePath)
			return
		}()
	}

	if err == io.EOF {
		err = nil
	}
	return
}