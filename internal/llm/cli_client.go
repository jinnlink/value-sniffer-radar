package llm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CLIClient struct {
	Command string
	Args    []string
	Timeout time.Duration
}

func (c CLIClient) Name() string { return "cli" }

func (c CLIClient) Complete(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(c.Command) == "" {
		return "", fmt.Errorf("missing cli command")
	}
	t := c.Timeout
	if t <= 0 {
		t = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.Command, c.Args...)
	cmd.Stdin = strings.NewReader(prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cli run failed: %w (stderr=%s)", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
