package main

import (
	"github.com/timperman/tail/plugin"
	"os"
)

func main() {
  port := os.Getenv("PLUGIN_PORT")
  if port == "" {
    port = ":8080"
  }

  root := os.Getenv("VOLUMES_ROOT")
  if root == "" {
    root = "/var/lib/tail-plugin"
  }

	plugin.Start(port, root)
}
