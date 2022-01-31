# Teleport Systems Challenge

This repository contains a solution for the [Teleport Systems Challenge](https://github.com/gravitational/careers/blob/main/challenges/systems/challenge.md) - a take-home project used by the Teleport team to assess their engineering candidates. This particular solution aims at the Level 5 of the challenge.

## System Design

You can find the RFD for the solution in [RFD.md](RFD.md). The file contains a lot of details about the API design, potential trade-offs and scope limits, etc.

## Testing scripts, etc

There are a few scripts and programs available to test the libraries shipped with this project:

### Tail

To demonstrate the usage of the `filestream` module, there is the `cmd/tail/tail.go` program. You can run it on any file like so:

```bash
go run cmd/tail/tail.go /path/to/file.log
```

The program is designed to tail the given file for 10 seconds and then time out. It could be stopped prematurely by pressing Ctrl+C (the signal should stop the process cleanly, propagating into the library).

### Containerize

To demonstrate the usage of the `containerize` module, there is a special script called `script/containerize.sh`. It accepts an arbitrary command as its arguments and executes it in an isolated environment while streaming its output to console. The command automatically times out after 30 seconds or could be stopped by a Ctrl+C. Since the command relies on Linux kernel APIs to implement process isolation, it runs in a privileged Docker container (to make it possible to develop on a Mac).

If you need access to some packages that are not available in the Docker container, feel free to adjust the `Dockerfile` used to build the environment.

Here are a few examples that may be useful for testing:

```bash
# Get some information about the system
./script/containerize.sh ps ax
./script/containerize.sh ls -lR /
./script/containerize.sh cat /etc/passwd

# Run a command that keeps streaming content until stopped (and demonstrates isolated networking)
./script/containerize.sh bash -c "ip link set lo up; ip addr list;ping 127.0.0.1"

# Test CPU limits
./script/containerize.sh sha256sum /dev/zero
```
