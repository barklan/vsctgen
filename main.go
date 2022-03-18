package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const start = `{
    "version": "2.0.0",
    "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": true,
        "panel": "shared",
        "showReuseMessage": true,
        "clear": true
    },
    "tasks": [`

const end = `
    ],
}
`

const backgroundTask = `
		{
			"label": "r/%s",
			"type": "shell",
			"command": "bash -i run.sh %s",
			"isBackground": true,
		},`

const normalTask = `
		{
			"label": "r/%s",
			"type": "shell",
			"command": "bash -i run.sh %s",
		},`

func runShCmds(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Panicln("failed to read run.sh", err)
	}

	r, err := regexp.Compile(`(\S*) *\(\) *({|\()`)
	if err != nil {
		log.Panicln(err)
	}

	matches := r.FindAllStringSubmatch(string(content), -1)

	cmds := make([]string, 0)
	for _, match := range matches {
		cmd := match[1]
		if !(string(cmd[0]) == "_" || cmd == "help") {
			cmds = append(cmds, match[1])
		}
	}
	return cmds
}

func main() {
	var runshFilepath, tasksFilepath string
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	runshFilepath = filepath.Join(path, "run.sh")
	tasksFilepath = filepath.Join(path, ".vscode", "tasks.json")
	if runshFilepath == "" || tasksFilepath == "" {
		log.Panicln("empty filepaths")
	}

	vscodeDir := strings.TrimSuffix(tasksFilepath, "/tasks.json")
	if err = os.Mkdir(vscodeDir, 0o755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			panic(err)
		}
	}

	cmds := runShCmds(runshFilepath)
	var opts string
	for _, cmd := range cmds {
		var task string
		if cmd == "up" || strings.Contains(cmd, "up:") {
			task = fmt.Sprintf(backgroundTask, cmd, cmd)
		} else {
			task = fmt.Sprintf(normalTask, cmd, cmd)
		}
		opts += task
	}

	taskfile := start + opts + end

	err = os.WriteFile(tasksFilepath, []byte(taskfile), 0o777)
	if err != nil {
		panic(err)
	}

	log.Println("tasks.json generated!")
}
