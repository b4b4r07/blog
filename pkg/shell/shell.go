package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
)

// New returns Shell instance
func New(command string, args ...string) Shell {
	return Shell{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Env:     map[string]string{},
		Command: command,
		Args:    args,
	}
}

// Shell represents shell command
type Shell struct {
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	OutputPrefix string
	ErrorPrefix  string
	Env          map[string]string
	Command      string
	Args         []string
	Dir          string
}

// Run runs shell command
func (s Shell) Run(ctx context.Context) error {
	command := s.Command
	if _, err := exec.LookPath(command); err != nil {
		return err
	}
	for _, arg := range s.Args {
		command += " " + arg
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Stderr = s.Stderr
	cmd.Stdout = s.Stdout
	cmd.Stdin = s.Stdin
	cmd.Dir = s.Dir
	for k, v := range s.Env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	return cmd.Run()
}

func (s Shell) RunWithPrefix(ctx context.Context, outpx, errpx string) error {
	command := s.Command
	if _, err := exec.LookPath(command); err != nil {
		return err
	}
	for _, arg := range s.Args {
		command += " " + arg
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Dir = s.Dir
	for k, v := range s.Env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	streamReader := func(scanner *bufio.Scanner, outputChan chan string, doneChan chan bool) {
		defer close(outputChan)
		defer close(doneChan)
		for scanner.Scan() {
			outputChan <- scanner.Text()
		}
		doneChan <- true
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stdoutOutputChan := make(chan string)
	stdoutDoneChan := make(chan bool)
	stderrScanner := bufio.NewScanner(stderr)
	stderrOutputChan := make(chan string)
	stderrDoneChan := make(chan bool)
	go streamReader(stdoutScanner, stdoutOutputChan, stdoutDoneChan)
	go streamReader(stderrScanner, stderrOutputChan, stderrDoneChan)

	stillGoing := true
	for stillGoing {
		select {
		case <-stdoutDoneChan:
			stillGoing = false
		case line := <-stdoutOutputChan:
			log.Printf("%s %s", outpx, line)
		case line := <-stderrOutputChan:
			log.Printf("%s %s", errpx, line)
		}
	}

	return cmd.Wait()
}

// RunCommand runs command with given arguments
func RunCommand(command string, args ...string) error {
	return New(command, args...).Run(context.Background())
}
