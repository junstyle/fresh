package runner

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func build() (string, bool) {
	buildLog("Building...")

	args := []string{"build"}
	args = append(args, buildArgs()...)
	args = append(args, "-o", buildPath(), root())
	cmd := exec.Command("go", args...)
	fmt.Printf("%#v", cmd.Args)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fatal(err)
	}
	flag.Parse()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		fatal(err)
	}

	io.Copy(os.Stdout, stdout)
	errBuf, _ := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	if err != nil {
		return string(errBuf), false
	}

	return "", true
}
