package tailcmd

import (
  "bufio"
  "encoding/json"
  "os/exec"
  "time"
)

type TailCmd struct {
  cmd *exec.Cmd
}

func TailPipe(file string, lines chan<- []byte) (*TailCmd, error) {
  cmd := exec.Command("tail", "-F", file)
  if err := cmd.Start(); err != nil {
    return nil, err
  }

  pipe, err := cmd.StdoutPipe()
  if err != nil {
    return nil, err
  }

  scanner := bufio.NewScanner(pipe)
  go scan(scanner, file, lines)

  return &TailCmd{ cmd: cmd, }, nil
}

func scan(scanner *bufio.Scanner, file string, lines chan<- []byte) {
  for scanner.Scan() {
    if bytes, err := json.Marshal(map[string]interface{}{ "file": file, "time": time.Now(), "line": scanner.Text() }); err == nil {
      lines <- bytes
    }
  }
}

func (tc *TailCmd) Stop() {
  tc.cmd.Process.Kill()
}
