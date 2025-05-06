package main

import (
	"goli-cli/cli"
	"goli-cli/utils/setUpUtils"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	defer func() {
		setUpUtils.UpdateCli()
	}()

	go func() {
		stopChan := make(chan os.Signal, 2)
		signal.Notify(stopChan,
			os.Interrupt,
			syscall.SIGINT,
			syscall.SIGKILL,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		<-stopChan
		setUpUtils.UpdateCli()
		os.Exit(0)
	}()

	cli.Execute()
}
