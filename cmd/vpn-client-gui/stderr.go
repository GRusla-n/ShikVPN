package main

import "os"

// defaultStderr preserves the original stderr before log.SetOutput redirects it.
var defaultStderr = os.Stderr
