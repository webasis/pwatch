package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Log struct {
	Time    time.Time
	IsError bool
	Content string
}

type Logger struct {
	IsError bool
	C       chan<- Log
	buf     *bytes.Buffer
}

func NewLogger(c chan<- Log, isError bool) *Logger {
	return &Logger{
		IsError: isError,
		C:       c,
		buf:     bytes.NewBuffer(nil),
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	n, err = l.buf.Write(p)

	for {
		line, e := l.buf.ReadString('\n')
		if len(line) > 0 {
			l.C <- Log{
				Time:    time.Now(),
				IsError: l.IsError,
				Content: line[:len(line)-1],
			}
		}
		if e != nil {
			return
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("pwatch prog {args}")
	}
	prog := os.Args[1]
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	C := make(chan Log, 1000)
	go func() {
		defer close(C)
		for {
			err := Cmd(C, prog, args...).Run()
			if err == nil {
				return
			} else {
				C <- Log{
					IsError: true,
					Content: "run: " + err.Error(),
					Time:    time.Now(),
				}
			}
		}
	}()

	for log := range C {
		prefix := ""
		suffix := "\x1b[0m\n"
		if log.IsError {
			prefix = "\x1b[31m"
		}
		fmt.Print(prefix, "[", log.Time.Format("15:04:05"), "] ", log.Content, suffix)
	}
}

func Cmd(ch chan<- Log, prog string, args ...string) *exec.Cmd {
	cmd := exec.Command(prog, args...)
	cmd.Stdout = NewLogger(ch, false)
	cmd.Stderr = NewLogger(ch, true)
	return cmd
}
