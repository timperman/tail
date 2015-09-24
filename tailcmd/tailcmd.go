package tailcmd

import (
  "bufio"
  "encoding/json"
  "log"
  "os/exec"
  "path/filepath"
  "time"
)

type TailCmd struct {
  cmd *exec.Cmd
}

func TailPipe(file string, lines chan<- []byte) (*TailCmd, error) {
  log.Printf("starting tail command to watch %s\n", file)
  cmd := exec.Command("/usr/bin/tail", "-F", file)

  pipe, err := cmd.StdoutPipe()
  if err != nil {
    log.Printf("error attaching to tail stdout: %v\n", err)
    return nil, err
  }

  errpipe, err := cmd.StderrPipe()
  if err != nil {
    log.Printf("error attaching to tail stdout: %v\n", err)
    return nil, err
  }

  if err := cmd.Start(); err != nil {
    log.Printf("error starting tail process: %v\n", err)
    return nil, err
  }

  scanner := bufio.NewScanner(pipe)
  go scan(scanner, filepath.Base(file), lines)

  errscan := bufio.NewScanner(errpipe)
  go func() {
    for errscan.Scan() {
      log.Printf("Error in tail process watching %s: %v\n", file, errscan.Text())
    }
  }()

  return &TailCmd{ cmd: cmd, }, nil
}

func scan(scanner *bufio.Scanner, file string, lines chan<- []byte) {
  for scanner.Scan() {
    if bytes, err := json.Marshal(map[string]interface{}{ "file": file, "time": time.Now(), "line": scanner.Text() }); err == nil {
      lines <- bytes
    } else {
      log.Printf("Error marshalling log line to JSON: %v\n", err)
    }
  }
}

func (tc *TailCmd) Stop() {
  tc.cmd.Process.Kill()
}
