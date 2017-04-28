package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"moby/cli/command"
	"moby/cli/debug"
	"github.com/stretchr/testify/assert"
)

func TestClientDebugEnabled(t *testing.T) {
	defer debug.Disable()

	cmd := newDockerCommand(&command.DockerCli{})
	cmd.Flags().Set("debug", "true")

	err := cmd.PersistentPreRunE(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "1", os.Getenv("DEBUG"))
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
}

func TestExitStatusForInvalidSubcommandWithHelpFlag(t *testing.T) {
	discard := ioutil.Discard
	cmd := newDockerCommand(command.NewDockerCli(os.Stdin, discard, discard))
	cmd.SetArgs([]string{"help", "invalid"})
	err := cmd.Execute()
	assert.EqualError(t, err, "unknown help topic: invalid")
}
