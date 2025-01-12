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

func RunCommand(cmd string, event func(event ShellEvent)) (cancel context.CancelFunc, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	eventChannel := make(chan ShellEvent)
	go func(ctx context.Context, outputChannel chan<- ShellEvent) {
		defer close(eventChannel)
		outputChannel <- ShellEvent{Type: "start", Data: cmd}

		command := exec.CommandContext(ctx, "sh", "-c", cmd)
		stdout, err := command.StdoutPipe()
		if err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
			return
		}

		if err := command.Start(); err != nil {
			outputChannel <- ShellEvent{Type: "error", Data: err.Error()}
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			outputChannel <- ShellEvent{Type: "output", Data: scanner.Text()}
		}

		if err := scanner.Err(); err != nil {
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
	}(ctx, eventChannel)

	go func() {
		for e := range eventChannel {
			event(e)
		}
	}()

	return cancel, nil
}

func RunCommandSync(cmd string) (output string, err error) {
	command := exec.Command("sh", "-c", cmd)
	stdout, err := command.Output()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}
