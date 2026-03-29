package term

import (
	"io"
	"os"
	"os/exec"
)

func OutputWriter() (io.Writer, func() error) {
	_, err := exec.LookPath("bat")
	if err != nil {
		return os.Stdout, func() error { return nil }
	}

	batCmd := exec.Command("bat", "--paging=never", "--language=md", "--plain")
	batStdin, err := batCmd.StdinPipe()
	if err != nil {
		return os.Stdin, func() error { return nil }
	}

	batCmd.Stdout = os.Stdout
	if err := batCmd.Start(); err != nil {
		return os.Stdin, func() error { return nil }
	}

	return batStdin, func() error {
		if err := batStdin.Close(); err != nil {
			return err
		}
		return batCmd.Wait()
	}
}
