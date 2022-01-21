package execute

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"

	log "github.com/asyrafduyshart/go-exec-engine/pkg/log"
	"github.com/linkedin/goavro"
	"github.com/xeipuuv/gojsonschema"
)

// Command ..
type Command struct {
	Name           string              `yaml:"name" validate:"required,alphanumunicode"`
	Protocol       string              `yaml:"protocol"`
	Target         string              `yaml:"target" validate:"required"`
	Exec           string              `yaml:"exec" validate:"required"`
	Type           string              `yaml:"type" validate:"required,oneof=http bash exec"`
	Validate       bool                `yaml:"validate"`
	Schema         string              `yaml:"schema"`
	SchemaType     string              `yaml:"schema-type" validate:"oneof=json avro"`
	Authentication bool                `yaml:"authentication"`
	ValidateClaim  map[string][]string `yaml:"validate-claim"`
}

// Execute ..
func Execute(command Command, data string) error {
	log.Info("Executed command %v", command.Exec)

	if command.Validate {
		if command.SchemaType == "avro" {
			if err := ValidateAvro(command, data); err != nil {
				log.Error("Trigger: \"%v\" will not be executed", command.Name)
				return err
			}
		} else if command.SchemaType == "json" {
			if err := ValidateJSON(command, data); err != nil {
				log.Error("Trigger: \"%v\" will not be executed", command.Name)
				return err
			}
		}
	}
	if command.Type == "exec" {
		s := strconv.Quote(string(data))
		exec := []string{"bash", "-c", "echo " + s + " |" + " " + command.Exec}
		log.Debug("Received Data %v", data)
		out, err := cmdExec(exec...)
		log.Info("Output: (%v)  \n%v", command.Name, out)
		if err != nil {
			log.Error("Error %v:", err)
			return err
		}
	} else if command.Type == "bash" {
		log.Debug("Received Data %v", data)
		out, err := scriptExec(command.Exec, data)
		log.Info("Output: (%v)  \n%v", command.Name, out)
		if err != nil {
			log.Error("Error %v:", err)
			return err
		}
	}

	return nil
}

func ValidateAvro(command Command, data string) error {
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

func ValidateJSON(command Command, data string) error {
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
		errorStr := ""
		for inte, desc := range result.Errors() {
			log.Error("- %s", desc)
			if inte > 0 {
				errorStr += "," + desc.Description()
			} else {
				errorStr += desc.Description()
			}
		}
		if len(result.Errors()) != 0 {
			return fmt.Errorf(errorStr)
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

func scriptExec(scriptName string, data string) (string, error) {
	err := os.Chmod(scriptName, 0744)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(scriptName)
	cmd.Env = os.Environ()

	var params map[string]interface{}
	json.Unmarshal([]byte(data), &params)
	reflectValue := reflect.ValueOf(params)

	// env variables definition
	param := reflectValue.MapRange()
	for param.Next() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", param.Key(), param.Value()))
	}

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("the err", err)
		return string(out), err
	}
	return string(out), nil
}
