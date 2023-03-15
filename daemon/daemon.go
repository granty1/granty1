package daemon

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const (
	// Env Daemon process env
	Env = "DAEMON_PROCESS=true"
)

var (
	// --daemon flag
	daemonFlag = flag.Bool("d", false, "run as daemon process")
)

func Run() {
	if !flag.Parsed() {
		flag.Parse()
	}

	if !(*daemonFlag) {
		return
	}

	Start()
}

func Start() {
	var childProc bool
	envs := os.Environ()
	for _, env := range envs {
		if env == Env {
			childProc = true
		}
	}

	if !childProc {
		// start fork
		cmd := exec.Command(os.Args[0], os.Args[1:]...)

		envs = append(envs, Env)
		log.Println("daemon proc start success!")
		cmd.Env = append(cmd.Env, envs...)
		if err := cmd.Start(); err != nil {
			error(err.Error())
		}
		// parent process exit success
		os.Exit(0)
	}
}

func error(content ...string) {
	fmt.Fprintln(os.Stdout, content)
	os.Exit(1)
}
