# Go Exec Engine

Go Exec Engine is an open source project that provides a framework for executing various types of commands. This includes support for executing shell commands, scripts, and HTTP requests. You can define these commands in a configuration file and control their execution through the API.

This project is useful for automating tasks, webhooks and microservices orchestration, and more.

## Features
* Support for multiple types of commands: HTTP, Bash and Exec.
* Support for different protocols: HTTP and Pub/Sub.
* JWT Authentication for HTTP requests.
* Command validation with JSON and Avro schemas.
* Ability to start, stop, and restart the execution engine.

## Installation

1. Clone the repository:

```bash
git clone https://github.com/asyrafduyshart/go-exec-engine.git
```

2. Change the working directory:

```bash
cd go-exec-engine
```

3. Install the project:

```bash
go install
```

## Usage

```bash
goexec [start|stop|restart] --config=<path_to_config_file>
```

## Configuration

The project uses a configuration file in YAML format. You can specify the path to the configuration file using the `--config` flag.

Here is an example of the configuration file:

```yaml
log_level: debug
jwks_url: "https://url/.well-known/jwks.json"
access_log:
commands:
  - name: "trigger"
    protocol: "pubsub"
    target: "dev-trigger-task"
    type: "exec"
    exec: "echo yomama"
    validate: false
    schema-type: "avro"
    schema: "test.avsc"
  - name: "exec"
    target: "exec-test"
    type: "exec"
    exec: "echo HelloWorld!"
    validate: true
    schema-type: "json"
    schema: "./schema/test.avsc"
  - name: "deployfrontend"
    protocol: "http"
    target: "/deploy/frontend"
    type: "bash"
    exec: "./deploy-frontend.sh"
    authentication: true
    validate-claim: 
      role: 
        - admin
        - employee
        - engineer
      scope: 
        - user:read_write
    validate: true
    schema-type: "json"
    schema: "./schema/deploy-frontend.json"
```

## Environment Variables

If you are using Pub/Sub from Google, set the following environment variable:

```bash
export GOOGLE_APPLICATION_CREDENTIALS="$(pwd)/googlekey.json"
```

## License

Go Exec Engine is open-source software licensed under the [MIT license](LICENSE).

## Author

Go Exec Engine was created by [Asyraf Duyshart](https://github.com/asyrafduyshart).