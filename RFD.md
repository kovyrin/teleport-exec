---
authors: Oleksiy Kovyrin (oleksiy@kovyrin.net)
state: draft
---

# RFD - Teleport Systems Challenge

## What

This document will be used to communicate and discuss design decisions for the
implementation of Teleport Systems Challenge by Oleksiy Kovyrin. See the [challenge page](https://github.com/gravitational/careers/blob/main/challenges/systems/challenge.md) for more details.

## Why

This exercise has two goals:

* It helps the Teleport team understand what to expect from Oleksiy as a developer, how he writes production code, how he reasons about API design and how he communicates when trying to understand a problem before solving it.
* It helps Oleksiy get a feel for what it would be like to work at Teleport, as this exercise aims to simulate their day-as-usual and expose Oleksiy to the type of work Teleport engineers are doing.

## Requirements

We need to build a client-server system that allows for remote execution of shell commands with per-command resource control (memory, CPU and disk IO) and isolation on network, process and filesystem levels using Linux kernel technologies like cgroups and kernel namespaces.

Here are the original requirements as described on the [challenge documentation page](https://github.com/gravitational/careers/blob/main/challenges/systems/challenge.md):


### Library

* Worker library with methods to start/stop/query status and get the output of a job.
* Library should be able to stream the output of a running job.
  * Output should be from start of process execution.
  * Multiple concurrent clients should be supported.
* Add resource control for CPU, Memory and Disk IO per job using cgroups.
* Add resource isolation for using PID, mount, and networking namespaces.

### API

* [GRPC](https://grpc.io) API to start/stop/get status/stream output of a running process.
* Use mTLS authentication and verify client certificate. Set up strong set of cipher suites for TLS and good crypto setup for certificates. Do not use any other authentication protocols on top of mTLS.
* Use a simple authorization scheme.

### Client

* CLI should be able to connect to worker service and start, stop, get status, and stream output of a job.

## Solution Details

### Security

A remote command execution service poses a number of challenges when it comes to security:

- The system needs to have a strong authentication and authorization system in place to ensure that only the people allowed to perform remote commands would be able to do so.

- The data coming through the system (commands, console output, etc) needs to be protected since it may contain sensitive information (credentials, customer data, etc).

- The host system running the server needs to be protected from potentially abusive and/or malicious users.

- Since the system is a multi-tenant environment, it may need some mechanisms to protect its users from noisy/malicious/abusive neighbors.

#### Authentication

We're going to use a dedicated certificate authority to issue server and client certificates and will use it to verify both the server and the client during the establishment of an mTLS connection to the service.

The following parameters will be used for all TLS sessions:
* TLS v1.3
* A secure default set of cipher suites:
    - Modern ciphers like Chacha20 and AES GCM
    - Key exchanges with support for perfect forward secrecy

##### Scope limits

CA, server and client certificate generation and management is out of scope for the project. A set of example certificates and keys will be checked into the repository with the project to allow for easier development of the system.

In an actual production system, certificates and keys would have to be configured using a configuration file or some other mechanism and their generation and management would depend on the underlying infrastructure managed by the operators of the service.

#### Authorization

Simple authorization tokens will be used by the client to gain access to the system after the initial mTLS connection is established. Each client could have its own set of access tokens and tokens could be revoked by the operators of the service if needed.

##### Scope limits

For this exercise, we're going to assume the following set of constraints:

* Tokens do not expire
* There is no user/token management infrastructure within the service (tokens will be hardcoded within a mock authorization service).
* The service will have two pre-defined roles:
  - `admin` - can submit any commands to the service, list all commands, and see output from any command
  - `user` - can submit any commands to the service, but can only list and see output from their own commands

In an actual production system, we'd probably want all authorization tokens to have some kind of TTL and would need a way to manage tokens on the server that did not require us to rebuild the server. A more flexible role system may be needed as well.

#### Data Protection

The data passing through the system needs to be protected in-flight (while being transferred between the client and the server) and at-rest (command logs need to be protected if persisted on disk). Additionally, users should only be able to see output from their own commands (unless an admin-level access token is used).

##### Scope limits

For this exercise, we're going to assume the following constraints:

* In-flight content protection will be implemented by encrypting all traffic.
* On-disk logs protection will be limited to file permissions (0600).
* Token-based authorization will be used to control who can see output from which commands (see the dedicated section above).

In a real production scenario we may want to encrypt on-disk logs (either do it ourselves or provide the operator with guidelines on how to achieve it via `dm-crypt`, encrypted EBS, or other technologies).

#### Container isolation

Since the system will be used to execute arbitrary commands from the users, we need to ensure full isolation of the user command from the underlying host OS. We're going to use Linux kernel namespaces to achieve isolation along the following dimensions:

* PID namespace - isolates user commands from the host OS process namespace to avoid leaking process information from the host into the container.

* Network namespace - isolates container networking stack from that of the host OS.

* Mount namespace - allows us to fully isolate command's filesystem from that of the host OS (and protect host OS `/proc` from being accessed by the user command to gain access to any privileged information on the host).

##### Scope limits

For this exercise, we're going to accept the following constraints:

* Networking namespace will be used to separate the user command from the host OS network, but we're not going to build any bridges, etc to connect container network to the host.

* We're going to reuse the host root filesystem instead of creating a separate clean FS for each container (using something like Alpine Root FS, etc).

#### Availability and resource limits

In a multi-tenant environment like the system in question, we need to ensure proper resource limits to reduce the potential impact of noisy or abusive clients on the underlying host OS.

We're going to use linux cgroup2 APIs to apply the following limits on each command:

* [Memory usage](https://facebookmicrosites.github.io/cgroup2/docs/memory-controller.html) – a static per-command limit will be applied on each container.

* [CPU usage](https://facebookmicrosites.github.io/cgroup2/docs/cpu-controller.html) - each command will be limited to a small constant amount of CPU resources, preventing it from negative affecting the host OS and other commands by consuming too much CPU.

* [Disk IO](https://facebookmicrosites.github.io/cgroup2/docs/io-controller.html) - each command will be limited by the amount of disk IO it could perform to limit the effects from a single command's high IO load on the host OS and other commands.

##### Scope limits

All limits will be statically hardcoded within the server, while in a production system we'd probably have some kind of configuration file for default limits and would expand the API to allow users to specify their own limits for each command (like Docker does).

In a production environment, we may want to introduce per-user resource limits, API call rate limits, etc to further protect the system from abuse.

### API

The following API methods will be available on the server:

* `Status` - returns some basic status information for the server, including the list of running and finished commands (for non-admin users only the processes owned by the user would be visible).
* `StartCommand` - starts a command on the server and returns a unique `command_id` value (a UUID string) used to manage the command in subsequent API calls.
* `StopCommand` - stops a given command (identified by a `command_id`).
* `CommandStatus` - returns current status information for a given command (identified by a `command_id`).
* `CommandOutput` - a streaming API for receiving console output (combined stdout+stderr) from a given command (identified by a `command_id`). If requested, the stream will continue until the command has finished (tail mode).

The detailed GRPC definition for the proposed API can be found within the [remote_exec.proto](remote_exec/remote_exec.proto) file.

### Library design

Most of concerns around containerization will be isolated within the `container_exec` library developed for this project, allowing us to have a very simple GRPC server around the library. This should make it easier to test the containerization code and contain most of the complexity (dealing with Linux APIs, process management, logs management, concurrency concerns, etc) the library.

In a production scenario, this approach would make it possible to potentially reuse the library in multiple different systems (see ContainerD core services or the Moby project and its usage within Docker).

#### Scope limits and design trade-offs

We're going to apply the following design decisions while building the library:

* Finished process state and logs will remain on the server and there will be no explicit way to delete those. It opens the server up to a number of DoS attacks and in a production scenario some kind of cleanup logic would need to be implemented (docker-like explicit delete command, time-based expiration, etc).

* To avoid storing all output in memory while allowing us to stream command output from the beginning, we follow the approach used by Docker: log into a file per container and stream from the file as needed. The file is never deleted (see above for clarifications).

* All processes started by the server share the same process group, so that we could kill them all as a group when killing the command.

#### Design details

##### Log streaming

When executing a new command, we will redirect its stdout/stderr to a command-specific log file. This makes it possible for us to stream log content to multiple clients at the same time using different stream positions and always have access to the whole command output no matter how long it is (limited by available disk space). The streaming of content from the end of the file may be implemented by using a library like [hpcloud/tail](https://github.com/hpcloud/tail) or built into the solution by either polling the `stat()` syscall results (and looking at the size of the file there) or by using the `inotify()` API for detecting file changes.

There will not be any log rotation, but in a production scenario we would probably add a limit on the size of the log and start rotating it eventually.

##### Process management

We're going to use the `exec.Command` Go API to start each command. Linux kernel namespaces will be enabled for each command via the `SysProcAttr` attribute.  For filesystem isolation (`/proc remounting, etc) and for applying Cgroups to the command process, we're going to rely on the [reexec](https://pkg.go.dev/github.com/docker/docker/pkg/reexec) package from Docker to allow us to apply the changes in the environment needed before a command is executed.

To avoid race conditions on reading and updating the status information for a command process, we're going to use a separate goroutine running `command.Wait()` to detect the moment when a process has finished and its status information has become available.

### CLI

The client binary will be used as the primary mechanism for accessing the service during the exercise. In an actual production environment, we could potentially build different clients (web ui, cli, etc) all relying on the same fundamental GRPC API.

The CLI will have the following commands available to the user:

* `server-status` - shows remote server status
* `run` - runs a remote command, waits for it to finish and streams it logs
* `start` - runs a remote command asynchronously and returns its command id
* `status` - shows the status of a given remote command
* `kill` - stops a remote command
* `logs` - shows remote command's console output (use `-tail` to follow the log)

The address of the server and the authorization token could be provided via command flags (e.g. `-addr string` and `-token string`) or via an environment variable.


#### Command examples

Example commands (all assume a token is provided via a TOKEN environment variable):

* Get remote server status (very simple call that could be used to test connection to the service and troubleshoot service connections)
  ```
  ./build/client server-status
  ```

* Run a remote command and stream its output to the console in a synchronous manner, show exit code at the end (convenience command, implemented by the client on top of lower-level APIs)
  ```
  ./build/client run hostname
  ```

* Start a command on a remote server asynchronously and get its command id
  ```
  ./build/client start cat /etc/hosts
  ```

* Get a status of a remote command with a given command id
  ```
  ./build/client status d2761bf6-c196-435b-96b4-d3560f82ee65
  ```

* Get console output of a a remote command with a given command id logged until now
  ```
  ./build/client logs d2761bf6-c196-435b-96b4-d3560f82ee65
  ```

* Get console output of a a remote command with a given command id and wait for more
  ```
  ./build/client logs -tail d2761bf6-c196-435b-96b4-d3560f82ee65
  ```

* Stop a remote command with a given command id
  ```
  ./build/client kill d2761bf6-c196-435b-96b4-d3560f82ee65
  ```
