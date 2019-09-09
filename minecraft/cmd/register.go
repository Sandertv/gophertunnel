package cmd

// commands holds a list of registered commands indexed by their name.
var commands = map[string]Command{}

// Register registers a command with a given alias. The command will be able to called with this alias,
// regardless of the actual name of the command.
func Register(alias string, command Command) {
	commands[alias] = command
}

// Command looks up a command by an alias. If found, the command and true are returned. If not, the returned
// command is nil and the bool is false.
func CommandByAlias(alias string) (Command, bool) {
	command, ok := commands[alias]
	return command, ok
}

// Commands returns a map of all registered commands indexed by the alias they were registered with.
func Commands() map[string]Command {
	return commands
}
