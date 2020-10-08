package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	execute "github.com/asyrafduyshart/go-exec-engine/pkg/execute"
	pubsub "github.com/asyrafduyshart/go-exec-engine/pkg/pubsub"

	log "github.com/asyrafduyshart/go-exec-engine/pkg/log"
	tools "github.com/asyrafduyshart/go-exec-engine/tools"

	"github.com/go-playground/validator"
	yaml "gopkg.in/yaml.v2"
)

// Config ...
type Config struct {
	AccessLog string            `yaml:"access_log"`
	LogLevel  string            `yaml:"log_level"`
	Command   []execute.Command `yaml:"commands,flow"`
}

const (
	// ProjectName ...
	ProjectName = "Go Exec Engine"
	// Version ...
	Version = "0.0.1"
	// PidFile ...
	PidFile = "goinx.pid"
)

var (
	configPath = flag.String("config", "config.yml", "Configuration Path")
	cmds       = []string{"start", "stop", "restart"}
)

func usage() {
	fmt.Printf("💖  %s %s\n", ProjectName, Version)
	fmt.Println("Author: Asyraf Duyshart")
	fmt.Println("Github: https://github.com/asyrafduyshart/go-exec-engine")
	fmt.Printf("\nUsage: goexec [start|stop|restart]\n")
	fmt.Println("Options:")
	fmt.Println("    --config\tConfiguration path")
	fmt.Println("    --help\tHelp info")
}

func startArgs() *Config {
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	if !tools.Contains(cmds, cmd) {
		usage()
		os.Exit(0)
	}

	// start goinx
	if cmd == cmds[0] {
		return start()
	}
	// stop goinx
	if cmd == cmds[1] {
		stop()
	}
	if cmd == cmds[2] {
		stop()
		return start()
	}

	return nil
}

func start() *Config {

	if tools.Exist(PidFile) {
		log.Warning("Goinx has bean started.")
		os.Exit(0)
	}

	conf := Config{}
	if pid := os.Getpid(); pid != 1 {
		err := ioutil.WriteFile(PidFile, []byte(strconv.Itoa(pid)), 0777)
		if err != nil {
			fmt.Println(err)
		}
	}

	flag.Usage = usage
	flag.Parse()

	var bytes []byte

	if mp := os.Getenv("CONFIG_SETTING"); mp != "" {
		bytes = []byte(mp)
	} else {
		result, err := ioutil.ReadFile(*configPath)
		if err != nil {
			log.Error("%v", err)
			os.Remove("goinx.pid")
			os.Exit(0)
		}
		bytes = result
	}

	err := yaml.Unmarshal([]byte(bytes), &conf)
	if err != nil {
		log.Error("%v", err)
		os.Remove("goinx.pid")
		os.Exit(0)
	}
	return &conf
}

func stop() {
	bytes, err := ioutil.ReadFile(PidFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	pid, err := strconv.Atoi(string(bytes))
	log.Info("Stopping " + strconv.Itoa(pid))

	fmt.Println()
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(0)
	os.Remove("goinx.pid")
}

func shutdownHook() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				os.Remove("goinx.pid")
				log.Info("Shutown Goinx.")
				os.Exit(0)
			default:
				log.Info("other", s)
			}
		}
	}()
}

func main() {

	ctx := context.Background()

	shutdownHook()

	conf := startArgs()
	log.Info("Start Goinx.")

	if conf.LogLevel == "debug" {
		log.LogLevelNum = 1
	}
	if conf.LogLevel == "info" {
		log.LogLevelNum = 2
	}
	if conf.LogLevel == "warn" {
		log.LogLevelNum = 3
	}
	if conf.LogLevel == "error" {
		log.LogLevelNum = 4
	}

	// log.Debug("Config Content: %v", conf)
	count := 0
	exitChan := make(chan int)
	for _, command := range conf.Command {
		var validate *validator.Validate
		validate = validator.New()
		err := validate.Struct(command)
		if err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				log.Error("Validation Error: %v", err)
			}
			for _, err := range err.(validator.ValidationErrors) {
				log.Error("Field in %v: %v %v %v", command.Name, err.StructField(), err.ActualTag(), err.Param())
				log.Error("Trigger: \"%v\" will not be executed", command.Name)
			}
		} else {
			go func(c execute.Command) {
				log.Info("Trigger(%v): \"%v\" is listening to target: %v", c.Type, c.Name, c.Target)
				err := pubsub.PullMsgs(ctx, "lido-white-label", c.Target, func(data string) {
					execute.Execute(c, data)
				})
				if err != nil {
					log.Error("Error in topic %v", c.Target)
					log.Error("Error: %v", err)
				} else {
					log.Info("Server now listning to pubsub topic %v", c.Target)
				}
				exitChan <- 0
			}(command)
			count++
		}
	}

	for i := 0; i < count; i++ {
		<-exitChan
	}

}
