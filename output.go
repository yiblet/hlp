package main

import (
	"io"
	"os"
	"os/exec"
)

// getOutputWriter returns an io.Writer that is either os.Stdout or a pipe into a call to bat
// depending on whether or not bat is available in the path.
func getOutputWriter() (io.Writer, func() error) {
	// Check if bat is available
	_, err := exec.LookPath("bat")
	if err != nil {
		// bat is not available, return os.Stdout
		return os.Stdout, func() error { return nil }
	}

	// bat is available, create a pipe to it
	batCmd := exec.Command("bat", "--paging=never", "--language=md", "--plain")
	batStdin, err := batCmd.StdinPipe()
	if err != nil {
		return os.Stdin, func() error { return nil }
	}

	// Set the output of bat to os.Stdout
	batCmd.Stdout = os.Stdout

	// Start the bat command
	if err := batCmd.Start(); err != nil {
		return os.Stdin, func() error { return nil }
	}

	// Return the bat stdin pipe as the writer
	return batStdin, func() error {
		if err := batStdin.Close(); err != nil {
			return err
		}
		return batCmd.Wait()
	}
}
