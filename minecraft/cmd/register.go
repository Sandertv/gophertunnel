package cmd

// commands holds a list of registered commands indexed by their name.
var commands = map[string]Command{}

// Register registers a command with its name and all aliases that it has. Any command with the same name or
// aliases will be overwritten.
func Register(command Command) {
	commands[command.name] = command
	for _, alias := range command.aliases {
		commands[alias] = command
	}
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
