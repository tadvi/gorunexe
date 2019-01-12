package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"ax/beep"
)

var timeout time.Duration

func init() {
	flag.DurationVar(&timeout, "timeout", 0, "timeout for program to run")
}

func gobuild() error {
	cmd := exec.Command("go", "build")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func startCmd(exeName string) {
	done := make(chan bool, 1)
	kill := make(chan bool, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			kill <- true
			time.Sleep(time.Second)
			log.Fatal("Received Ctrl-C. Exit.")
		}
	}()

	fmt.Println("Running file", exeName, flag.Args())
	cmd := exec.Command(exeName, flag.Args()...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	go func() {
		cmd.Run()
		done <- true
	}()

	if timeout > 0 {
		go func() {
			<-time.After(timeout)
			fmt.Printf("Timeout %s. Process kill.\n", timeout)
			beep.Alert()
			kill <- true
		}()
	}

	select {
	case <-done:
		break
	case <-kill:
		cmd.Process.Kill()
	}
}

func main() {
	flag.Parse()

	if err := gobuild(); err != nil {
		log.Fatal(err)
	}

	match, err := filepath.Glob("*.exe")
	if err != nil {
		log.Println(err)
		return
	}

	switch m := len(match); {
	case m == 0:
		log.Fatal("No executable in this folder")
	case m == 1:
		startCmd(match[0])
	case m > 1:
		log.Fatal("Too many executables in this folder")
	}
}
