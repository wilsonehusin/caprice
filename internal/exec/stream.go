package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/wilsonehusin/caprice/scribe"
)

type loggedStream struct {
	realStream *os.File
	logFile    *os.File
}

func (s *loggedStream) Write(p []byte) (int, error) {
	w := io.MultiWriter(s.realStream, s.logFile)

	return w.Write(p)
}

func (s *loggedStream) Close() error {
	return s.logFile.Close()
}

type ExecStreams struct {
	stdout *loggedStream
	stderr *loggedStream
}

func NewExecStreams(logDir string) (*ExecStreams, error) {
	scribeRuntimeID := scribe.RuntimeID()

	e := &ExecStreams{}

	if logDir == "" {
		logDir = os.TempDir()
	}

	stdoutLogFile, err := os.Create(path.Join(
		logDir,
		fmt.Sprintf("%s.out", scribeRuntimeID),
	))
	if err != nil {
		return nil, err
	}
	e.stdout = &loggedStream{
		realStream: os.Stdout,
		logFile:    stdoutLogFile,
	}

	stderrLogFile, err := os.Create(path.Join(
		logDir,
		fmt.Sprintf("%s.err", scribeRuntimeID),
	))
	if err != nil {
		return nil, err
	}
	e.stderr = &loggedStream{
		realStream: os.Stderr,
		logFile:    stderrLogFile,
	}

	return e, nil
}

func (e *ExecStreams) Close() error {
	errs := []error{}
	if err := e.stdout.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := e.stderr.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("closing streams: %v", errs)
	}
	return nil
}

func (e *ExecStreams) AttachCmd(cmd *exec.Cmd) {
	cmd.Stdout = e.stdout
	cmd.Stderr = e.stderr
}
