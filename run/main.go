package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
)

func main() {
	args := []string{"--api-sock", "/tmp/firecracker.sock"}
	cmd := exec.Command("./firecracker", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	scannerr := bufio.NewScanner(stderr)
	go func() {
		for scannerr.Scan() {
			fmt.Printf("%s\n", scannerr.Text())
		}
	}()

	fmt.Printf("Waiting for firecracker to finish...")
	err = cmd.Wait()
	fmt.Printf("Firecracker finished with error: %v", err)
}
