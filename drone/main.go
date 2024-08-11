package main

import (
	"github.com/boar-d-white-foundation/drone/cli"
	_ "go.uber.org/automaxprocs"
)

func main() {
	cli.Run("drone-start", startDrone)
}
