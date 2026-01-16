package command

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/agent/runtime/plugins"
	"github.com/intrik8-labs/smidr/internal/logging"
)

type CommandPlugin struct{}

func NewCommandPlugin() *CommandPlugin {
	return &CommandPlugin{}
}

func (p *CommandPlugin) Execute(ctx context.Context, j job.Job, req plugins.CommandRequest) (*job.JobResponse, error) {
	logger := logging.FromContext(ctx).With("plugin", "command", "job_id", j.ID)

	logger.Debug("Starting command execution")

	// Validate required fields
	if req.Command == "" {
		return &job.JobResponse{
			Status:       "failed",
			Message:      "Command is required",
			ErrorMessage: "command field cannot be empty",
		}, fmt.Errorf("command is required")
	}

	// Set defaults
	if req.WorkDir == "" {
		req.WorkDir = "."
	}
	if req.Shell == "" {
		req.Shell = "/bin/sh"
	}

	// Apply timeout if specified
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	startTime := time.Now()

	logger.Info("Executing command",
		"command", req.Command,
		"args", req.Args,
		"workdir", req.WorkDir,
		"shell", req.Shell,
		"timeout", req.Timeout)

	// Build the command
	var cmd *exec.Cmd
	if len(req.Args) > 0 {
		// Direct command with args
		cmd = exec.CommandContext(ctx, req.Command, req.Args...)
	} else {
		// Shell command
		cmd = exec.CommandContext(ctx, req.Shell, "-c", req.Command)
	}

	cmd.Dir = req.WorkDir

	// Set environment variables
	if len(req.EnvVars) > 0 {
		env := cmd.Environ()
		for k, v := range req.EnvVars {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// Execute command
	var stdout, stderr bytes.Buffer
	if req.CaptureOutput {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	execTime := time.Since(startTime)

	// Prepare response
	response := &job.JobResponse{
		Metadata: map[string]string{
			"execution_time": execTime.String(),
			"command":        req.Command,
			"workdir":        req.WorkDir,
			"exit_code":      "0",
		},
	}

	if err != nil {
		exitCode := "unknown"
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = fmt.Sprintf("%d", exitErr.ExitCode())
		}
		response.Metadata["exit_code"] = exitCode

		logger.Error("Command execution failed",
			"error", err,
			"exit_code", exitCode,
			"execution_time", execTime,
			"stderr", stderr.String())

		response.Status = "failed"
		response.Message = fmt.Sprintf("Command failed with exit code %s", exitCode)
		response.ErrorMessage = err.Error()

		if req.CaptureOutput && stderr.Len() > 0 {
			response.Metadata["stderr"] = stderr.String()
		}

		return response, err
	}

	logger.Info("Command executed successfully",
		"execution_time", execTime,
		"output_size", stdout.Len())

	response.Status = "completed"
	response.Message = fmt.Sprintf("Command executed successfully in %s", execTime)

	if req.CaptureOutput {
		if stdout.Len() > 0 {
			response.Metadata["stdout"] = stdout.String()
		}
		if stderr.Len() > 0 {
			response.Metadata["stderr"] = stderr.String()
		}
	}

	return response, nil
}
