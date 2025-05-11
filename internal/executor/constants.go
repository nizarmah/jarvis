package executor

var (
	// Commands is the list of commands that the executor can execute.
	Commands = []string{
		"pause_video",
		"play_video",
	}
	// CommandsHumanReadable is the human readable version of the commands.
	CommandsHumanReadable = map[string]string{
		"pause_video": "pause the YouTube video",
		"play_video":  "play the YouTube video",
	}
)
