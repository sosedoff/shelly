# Shelly

Experimental remote shell runner written in Go

## Overview

Attempt to create a simple TCP server that runs bash commands locally from remote
connection. Since Go binary file is compiled it makes is very easy to deploy. It
is still in very early development, so lots of things might change down the road.

## Install

There are no external dependencies. Clone repository and you should be good to go:

```
git clone git://github.com/sosedoff/shelly.git
cd shelly
go build main.go
```

## Usage

Execute a shell (bash) command:

```go
cmd := Exec("ping -c 5 google.com")
if (cmd.Success()) {
  cmd.Print()
}
```

Result (Command struct) will include the following:

- `Command`    - Original command (string)
- `ExitStatus` - Execution exit status (int)
- `Output`     - Execution output (string)
- `TimeStart`  - Start timestamp (time.Time)
- `TimeFinish` - Completion timestamp (time.Time)
- `Duration`   - Execution duration (time.Duration)

To transmit over JSON, run:

```
cmd.ToJson()
```