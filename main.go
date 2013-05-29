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
  BIND_ADDR   = "0.0.0.0:20000"
  BUFFER_SIZE = 1024
)

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

func HandleConnection(socket net.Conn) {
  buffer := make([]byte, BUFFER_SIZE)

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

func main() {
  fmt.Printf("Starting server on %s\n", BIND_ADDR)

  server, err := net.Listen("tcp", BIND_ADDR)
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