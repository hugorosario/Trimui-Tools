package shell

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
)

type ShellEvent struct {
	Type string
	Data string
}

func RunCommandAsync(cmd string, event func(event ShellEvent)) (stdinChannel chan string, cancel context.CancelFunc, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	eventChannel := make(chan ShellEvent)
	stdinChannel = make(chan string)

	go func(ctx context.Context, outputChannel chan<- ShellEvent, stdinChannel <-chan string) {
		defer close(eventChannel)
		outputChannel <- ShellEvent{Type: "start", Data: cmd}

		command := exec.CommandContext(ctx, "sh", "-c", cmd)
		stdout, err := command.StdoutPipe()
		if err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
			return
		}

		stderr, err := command.StderrPipe()
		if err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
			return
		}

		stdinPipe, err := command.StdinPipe()
		if err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
			return
		}

		if err := command.Start(); err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
			return
		}

		go func() {
			for data := range stdinChannel {
				_, err := fmt.Fprint(stdinPipe, data)
				if err != nil {
					outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
					return
				}
			}
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			data := scanner.Text()
			outputChannel <- ShellEvent{Type: "output", Data: data}
		}

		if err := scanner.Err(); err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
		}

		errscanner := bufio.NewScanner(stderr)
		for errscanner.Scan() {
			data := errscanner.Text()
			outputChannel <- ShellEvent{Type: "error", Data: data}
		}

		if err := errscanner.Err(); err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
		}

		if err := command.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode := exitError.ExitCode()
				outputChannel <- ShellEvent{Type: "exit", Data: fmt.Sprintf("%d", exitCode)}
			} else {
				outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
			}
		} else {
			outputChannel <- ShellEvent{Type: "exit", Data: "0"}
		}

		outputChannel <- ShellEvent{Type: "end", Data: cmd}
	}(ctx, eventChannel, stdinChannel)

	go func() {
		for e := range eventChannel {
			event(e)
		}
	}()

	return stdinChannel, cancel, nil
}

func RunCommandSync(cmd string) (output string, err error) {
	command := exec.Command("sh", "-c", cmd)
	stdout, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}
