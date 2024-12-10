package bash

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/jolt9dev/go-env"
	"github.com/jolt9dev/go-exec"
	"github.com/jolt9dev/go-fs"
	"github.com/jolt9dev/go-platform"
	"github.com/jolt9dev/go-xstrings"
)

var wslInstalled = false

func init() {
	exec.Register("bash", &exec.Executable{
		Name:     "bash",
		Variable: "BASH_PATH",
		Windows: []string{
			"${ProgramFiles}\\Git\\bin\\bash.exe",
			"${ProgramFiles}\\Git\\usr\\bin\\bash.exe",
			"${ProgramFiles(x86)}\\Git\\bin\\bash.exe",
			"${ProgramFiles(x86)}\\Git\\usr\\bin\\bash.exe",
			"${SystemRoot}\\System32\\bash.exe",
		},
		Linux: []string{
			"/bin/bash",
			"/usr/bin/bash",
		},
	})

	if platform.IsWindows() {
		drive := env.Get("SystemRoot")
		if drive == "" {
			drive = "C:\\Windows"
		}

		fp := filepath.Join(drive, "System32", "wsl.exe")

		fi, err := fs.Stat(fp)
		wslInstalled = err == nil && !fi.IsDir()
	}
}

// Returns the path to the bash executable or an empty string
func Which() string {
	exe, _ := exec.Find("bash")
	return exe
}

// Returns the path to the bash executable or the default
// which is the name of the executable without a path or
// extension.
func WhichOrDefault() string {
	exe, _ := exec.Find("bash")
	if exe == "" {
		return "bash"
	}

	return exe
}

// Creates a new bash command with the given arguments
// using vardiac arguments
//
// Example:
//
//	bash.New("--norc", "-e", "-o", "pipefail", "-c", "echo hello").Run()
func New(args ...string) *exec.Cmd {
	return exec.New(WhichOrDefault(), args...)
}

// Creates a new bash command with the given arguments
// using a single string
//
// Example:
//
//	bash.Command("--norc -e -o pipefail -c 'echo hello'").Run()
func Command(args string) *exec.Cmd {
	return exec.New(WhichOrDefault(), exec.SplitArgs(args)...)
}

// Creates a new bash command with the given script file
//
// Example:
//
//	bash.File("script.sh").Run()
func File(file string) *exec.Cmd {
	args := []string{"-noprofile", "--norc", "-e", "-o", "pipefail"}
	exe := WhichOrDefault()
	if wslInstalled {
		if xstrings.HasSuffixFold("System32\\bash.exe", exe) {
			f, err := filepath.Abs(file)
			if err == nil {
				file = f
			}

			file = "/mnt/" + string(unicode.ToLower(rune(file[0]))) + file[2:]
			file = filepath.ToSlash(file)
		}
	}

	args = append(args, file)
	return exec.New(exe, args...)
}

// Creates a new bash command with the given inline script
// or file. However, the file must have a .sh extension
// and be on a single line.
//
// Example:
//
//	bash.Script(`apt install -y age \
//	  curl \
//	  zip`).WithCwd("/path/to/dir").Run()
//	bash.Script("/path/to/script.sh").Output()
func Script(script string) *exec.Cmd {
	if !strings.ContainsAny(script, "\n") {
		script = strings.TrimSpace(script)

		if strings.HasSuffix(script, ".sh") {
			return File(script)
		}
	}

	args := []string{"-noprofile", "--norc", "-e", "-o", "pipefail", "-c", script}
	return exec.New(WhichOrDefault(), args...)
}

// Run a new bash inline script or file.
// When using a file, the file must have a .sh extension
// and be on a single line.
// Run will set stdout and stderr to inherit and not
// capture the output.
//
// Example:
//
//	bash.Run(`apt install -y age \
//	  curl \
//	  zip`).Run()
//	bash.Run("/path/to/script.sh")
func Run(script string) (*exec.PsOutput, error) {
	return Script(script).Run()
}

// Output a new bash inline script or file.
// When using a file, the file must have a .sh extension
// and be on a single line.
// Output will set stdout and stderr to piped and captures
// the standard output and error streams
//
// Example:
//
//	 bash.Output(`apt install -y age \
//		  curl \
//		  zip`).Run()
//	 out, err := bash.Output("/path/to/script.sh")
//	 if err != nil || out.Code != 0 {
//	 // handle error
//	 }
func Output(script string) (*exec.PsOutput, error) {
	return Script(script).Output()
}
