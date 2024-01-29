package main

import (
	"io"
	"os"

	"github.com/alexflint/go-arg"
)

func WriteHelp(config any, output io.Writer) error {
	parser, err := arg.NewParser(arg.Config{}, config)
	if err != nil {
		return err
	}
	parser.WriteHelp(output)
	return nil
}

func copyFile(src, dst string) error {
	// Open the source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Sync the written content to the disk
	err = dstFile.Sync()
	if err != nil {
		return err
	}

	// Set the same permissions as the source file
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	err = os.Chmod(dst, srcFileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}
