package main

import (
  "fmt"
  "os"
  "os/exec"
  "syscall"
  "time"
  "bytes"
  "encoding/json"
  "net"
  "strings"
)

const (
  SHELLY_VERSION = "0.1.0"
  SHELLY_BIND    = "0.0.0.0:20000"
  SHELLY_BUFFER  = 4096
)

var authToken string

type Command struct {
  Command    string         `json:"command"`
  ExitStatus int            `json:"exit_status"`
  Output     string         `json:"output"`
  TimeStart  time.Time      `json:"time_start"`
  TimeFinish time.Time      `json:"time_finish"`
  Duration   time.Duration  `json:"duration"`
}

func (cmd *Command) Run(command string) {
  var output bytes.Buffer
  shell := exec.Command("bash", "-c", command)

  shell.Stdout  = &output
  shell.Stderr  = &output
  cmd.Command   = command
  cmd.TimeStart = time.Now()

  err := shell.Start()
  if (err != nil) {
    fmt.Println("Execution failed:", err)
  }

  err = shell.Wait()

  cmd.TimeFinish = time.Now()
  cmd.Duration   = cmd.TimeFinish.Sub(cmd.TimeStart)

  if msg, ok := err.(*exec.ExitError); ok {
    cmd.ExitStatus = msg.Sys().(syscall.WaitStatus).ExitStatus()
  } else {
    cmd.ExitStatus = 0
  }

  cmd.Output = string(output.Bytes())
}

func (cmd *Command) ToJson() (s string) { 
  buff, err := json.Marshal(cmd)

  if (err != nil) {
    s = ""
    return
  }

  return string(buff)
}

func (cmd *Command) Print() {
  fmt.Println("Command:   ", cmd.Command)
  fmt.Println("ExitStatus:", cmd.ExitStatus)
  fmt.Println("Duration:  ", cmd.Duration)
  fmt.Println("Output:    ", cmd.Output)
}

func (cmd *Command) Success() bool {
  return cmd.ExitStatus == 0
}

func Exec(str string) *Command {
  var command *Command
  command = new(Command)
  command.Run(str)

  return command
}

func WriteWelcome(socket net.Conn) error {
  msg := fmt.Sprintf("Shelly v%s\n", SHELLY_VERSION)
  _, err := socket.Write([]byte(msg))

  return err
}

func ConnectionValid(socket net.Conn, buffer []byte) bool {
  num, err := socket.Read(buffer)

  if err != nil {
    return false
  }

  token := strings.TrimSpace(string(buffer[0:num]))

  return token == authToken
}

func HandleConnection(socket net.Conn) {
  buffer := make([]byte, SHELLY_BUFFER)

  /* Verify client token */
  if !ConnectionValid(socket, buffer) {
    fmt.Println("Client verification failed")
    socket.Close()
    return
  }

  /* Write welcome message */
  if WriteWelcome(socket) != nil {
    fmt.Println("Failed to welcome connection")
    socket.Close()
    return
  }

  for {
    num, err := socket.Read(buffer)
  
    if err != nil {
      fmt.Println("Read error:", err.Error())
      break
    }

    cmd := strings.TrimSpace(string(buffer[0:num]))

    if (len(cmd) == 0) {
      continue
    }

    if (cmd == "!done") {
      break
    }

    fmt.Println("Executing:", cmd)
    result := Exec(cmd)
    _, err = socket.Write([]byte(result.ToJson()))
  }

  fmt.Println("Client connection closed")
  socket.Close()
}

func EnvVarDefined(name string) bool {
  result := os.Getenv(name)
  return len(result) > 0
}

func main() {
  if !EnvVarDefined("SHELLY_TOKEN") {
    fmt.Println("SHELLY_TOKEN environment variable required")
    os.Exit(1)
  }

  authToken = os.Getenv("SHELLY_TOKEN")
  bindAddr := SHELLY_BIND

  if EnvVarDefined("SHELLY_BIND") {
    bindAddr = os.Getenv("SHELLY_BIND")
  }

  fmt.Printf("Starting server on %s\n", bindAddr)

  server, err := net.Listen("tcp", bindAddr)
  if err != nil {
    fmt.Println("Error:", err.Error())
    os.Exit(1)
  }

  for {
    socket, err := server.Accept()
    if err != nil {
      fmt.Println("Accept error:", err.Error())
      return
    }
    
    go HandleConnection(socket)
  }
}