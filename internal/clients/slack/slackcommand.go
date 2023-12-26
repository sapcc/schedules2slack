package slack

import (
	"github.com/sapcc/pulsar/pkg/auth"
	"github.com/sapcc/pulsar/pkg/util"
	"github.com/slack-go/slack"
)

var availableCommands = make([]CommandFactory, 0)

// Command is the interface for slack bot commands.
type Command interface {

	// Init maybe used for initializing the command.
	Init() error

	// Describe returns a short description of the command.
	Describe() string

	// Keywords is a list of strings which trigger this command.
	Keywords() []string

	// IsDisabled can be used to (temporarily) disable a command.
	IsDisabled() bool

	// RequiredUserRole returns the UserRole required to run the command.
	// Should at least return auth.UserRoles.Base .
	RequiredUserRole() auth.UserRole

	// Run executes the command and returns the response or an error.
	Run(originalMsg *slack.Msg) (*slack.Msg, error)
}

type CommandFactory func() Command

// RegisterCommand registers a new command if not already done.
func RegisterCommand(factory CommandFactory) {
	if factory == nil {
		return
	}

	f := factory()
	for _, knownCommand := range availableCommands {
		// Return here if the command is already registered (equal description or keywords or if it is marked as disabled.
		if knownCommand().Describe() == f.Describe() || util.IsSlicesEqual(knownCommand().Keywords(), f.Keywords()) || f.IsDisabled() {
			return
		}
	}
	availableCommands = append(availableCommands, factory)
}
