package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
)

type Args struct {
	Command string
}

type Result struct {
	Stdout string
	Stderr string
	Code   int
}

type Root int

func (t *Root) Run(args *Args, reply *Result) error {
	log.Println(args.Command)
	cmd := exec.Command("/bin/sh", "-c", args.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	var code int
	if exitError, ok := err.(*exec.ExitError); ok {
		code = exitError.ExitCode()
	}

	*reply = Result{Stdout: stdout.String(), Stderr: stderr.String(), Code: code}
	log.Printf("reply: %v", reply)
	return nil
}

func main() {
	switch os.Args[1] {
	case "client":
		client, err := rpc.Dial("tcp", os.Args[2])
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			var reply = &Result{}
			err = client.Call("Root.Run", &Args{Command: line}, reply)
			if err != nil {
				panic(err)
			}
			log.Printf("reply: %v", reply)
		}

	case "client-oneshot":
		client, err := rpc.Dial("tcp", os.Args[2])
		if err != nil {
			panic(err)
		}

		var reply = &Result{}
		err = client.Call("Root.Run", &Args{Command: os.Args[3]}, reply)
		if err != nil {
			panic(err)
		}

		fmt.Fprint(os.Stdout, reply.Stdout)
		fmt.Fprint(os.Stderr, reply.Stderr)
		os.Exit(reply.Code)

	case "server":
		arith := new(Root)
		rpc.Register(arith)
		l, err := net.Listen("tcp", os.Args[2])
		if err != nil {
			log.Fatal(err)
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err)
			}
			go rpc.ServeConn(conn)
		}
	default:
		panic("invalid argument")
	}
}
