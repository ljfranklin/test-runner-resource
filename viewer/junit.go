package viewer

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/ljfranklin/test-runner-resource/models"
)

// go:generate counterfeiter . Junit

type Junit interface {
	PrintSummary(models.Summary) error
}

type JunitCLI struct {
	OutputWriter io.Writer
	ResultsDir   string
}

func (j JunitCLI) PrintSummary(summary models.Summary) error {
	xmlFiles, err := filepath.Glob(filepath.Join(j.ResultsDir, "*.xml"))
	if err != nil {
		return fmt.Errorf("unable to glob for files: %s", err)
	}
	if len(xmlFiles) == 0 {
		return fmt.Errorf("found no .xml files in results dir '%s'", j.ResultsDir)
	}
	args := []string{
		"-o", summary.Type,
		"-l", fmt.Sprintf("%d", summary.Limit),
	}
	args = append(args, xmlFiles...)

	cmd := exec.Command("junit-viewer", args...)
	cmd.Stdout = j.OutputWriter
	cmd.Stderr = j.OutputWriter

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to print summary: %s", err)
	}
	return nil
}
