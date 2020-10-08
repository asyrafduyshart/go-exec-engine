package execute

import (
	"os/exec"

	log "github.com/asyrafduyshart/go-exec-engine/pkg/log"
)

// Command ..
type Command struct {
	Target string `yaml:"target"`
	Exec   string `yaml:"exec"`
}

// Execute ..
func Execute(command Command, data string) {
	log.Info("Executed command %v", command.Exec)
	exec := []string{"bash", "-c", "echo " + data + " |" + " " + command.Exec}
	out, err := cmdExec(exec...)
	if err != nil {
		log.Error("Error %v:", err)
	}
	log.Debug("Received Data %v", data)
	log.Info("Output:  \n%v", out)
}

func cmdExec(args ...string) (string, error) {

	baseCmd := args[0]
	cmdArgs := args[1:]

	log.Debug("Exec: %v", args)

	cmd := exec.Command(baseCmd, cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
