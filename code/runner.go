package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "code-user/main.go")
	var out, stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Stdout = &out
	stdinPipe, err := cmd.StdinPipe()
	//根据测试的输入案例输出结果 和标准输出的结果比对
	if err != nil {
		log.Fatalln(err)
	}
	io.WriteString(stdinPipe, "23 11\n")
	if err := cmd.Run(); err != nil {
		log.Fatalln(err, stderr.String())
	}
	fmt.Println(out.String())

	println(out.String() == "34\n")
}
