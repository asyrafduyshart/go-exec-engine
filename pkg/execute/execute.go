package execute

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	log "github.com/asyrafduyshart/go-exec-engine/pkg/log"
	"github.com/linkedin/goavro"
	"github.com/xeipuuv/gojsonschema"
)

// Command ..
type Command struct {
	Name       string `yaml:"name" validate:"required,alphanumunicode"`
	Target     string `yaml:"target" validate:"required"`
	Exec       string `yaml:"exec" validate:"required"`
	Type       string `yaml:"type" validate:"required,oneof=http bash"`
	Validate   bool   `yaml:"validate"`
	Schema     string `yaml:"schema"`
	SchemaType string `yaml:"schema-type" validate:"oneof=json avro"`
}

// Execute ..
func Execute(command Command, data string) {
	log.Info("Executed command %v", command.Exec)

	if command.Validate {
		if command.SchemaType == "avro" {
			if err := validateAvro(command, data); err != nil {
				log.Error("Trigger: \"%v\" will not be executed", command.Name)
				return
			}
		} else if command.SchemaType == "json" {
			if err := validateJSON(command, data); err != nil {
				log.Error("Trigger: \"%v\" will not be executed", command.Name)
				return
			}
		}
	}

	if command.Type == "bash" {
		exec := []string{"bash", "-c", "echo " + data + " |" + " " + command.Exec}
		out, err := cmdExec(exec...)
		if err != nil {
			log.Error("Error %v:", err)
		}
		log.Debug("Received Data %v", data)
		log.Info("Output:  \n%v", out)
	} else if command.Type == "http" {

	}
}

func validateAvro(command Command, data string) error {
	content, err := ioutil.ReadFile(command.Schema)
	if err != nil {
		log.Error("Error readfile: %v", err)
		return err
	}
	text := string(content)
	codec, err := goavro.NewCodec(text)
	if err != nil {
		log.Error("Error avro codec: %v", err)
		return err
	}

	decoded, _, err := codec.NativeFromTextual([]byte(data))
	if err != nil {
		log.Error("NativeFromTextual error: %v", err)
		return err
	}
	log.Info("Data succesfully decoded: %v", decoded)
	return nil
}

func validateJSON(command Command, data string) error {
	content, err := ioutil.ReadFile(command.Schema)
	if err != nil {
		log.Error("Error readfile: %v", err)
		return err
	}

	schemaLoader := gojsonschema.NewStringLoader(string(content))
	documentLoader := gojsonschema.NewStringLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Error("Error json codec: %v", err)
		return err
	}
	if result.Valid() {
		log.Info("Data is valid & verified")
	} else {
		log.Error("The document is not valid. see errors :")
		for _, desc := range result.Errors() {
			log.Error("- %s", desc)
		}
		if len(result.Errors()) != 0 {
			return fmt.Errorf("JSON validation failed")
		}
	}

	return nil
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
