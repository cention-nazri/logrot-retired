package logrot

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"testing"
	"time"
)

func remove(t *testing.T, files ...string) {
	if t.Failed() {
		return
	}

	for _, f := range files {
		os.Remove(f)
	}
}

func readFile(t *testing.T, name string) string {
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Printf("%v\n%s", err, debug.Stack())
		t.Fatalf("error reading %s\n", name)
	}
	return string(buf)
}

func TestWriteTo(t *testing.T) {
	defer remove(t, "log.txt")
	os.Remove("log.txt")
	log.SetFlags(0)

	defer WriteTo("log.txt").Close()

	want := "some log\n"
	log.Println("some log")
	fmt.Fprintln(os.Stdout, "This is from stdout")
	fmt.Fprintln(os.Stderr, "This is from stderr")
	got := readFile(t, "log.txt")

	if got != want {
		t.Errorf("\nwant: '%s'\n got: '%s'", want, got)
	}
}

func TestWriteAllTo(t *testing.T) {
	os.Remove("log.txt")
	log.SetFlags(0)
	stdout := os.Stdout
	stderr := os.Stderr

	defer WriteAllTo("log.txt").Close()

	want := "some log\nThis is from stdout\nThis is from stderr\n"

	log.Println("some log")
	fmt.Fprintln(os.Stdout, "This is from stdout")
	fmt.Fprintln(os.Stderr, "This is from stderr")

	os.Stdout = stdout
	os.Stderr = stderr

	got := readFile(t, "log.txt")
	if got != want {
		t.Errorf("\nwant: '%s'\n got: '%s'", want, got)
	}
}

func TestRotate(t *testing.T) {
	os.Remove("log.txt")
	os.Remove("log.txt.old")
	log.SetFlags(0)
	stdout := os.Stdout
	stderr := os.Stderr
	defer remove(t, "log.txt", "log.txt.old")

	rl := WriteTo("log.txt")
	defer rl.Close()
	log.Println("some log")
	err := os.Rename("log.txt", "log.txt.old")
	if err != nil {
		os.Stdout = stdout
		os.Stderr = stderr
		t.Fatalf("TestRotate(): %v\n", err)
	}
	log.Println("new filename")

	got := readFile(t, "log.txt.old")
	want := "some log\nnew filename\n"

	if got != want {
		t.Errorf("log file renamed failed\nwant: '%s'\n got: '%v'", want, got)
	}

	rl.rotate()
	want = "This is after rotation\n"
	log.Printf(want)

	os.Stdout = stdout
	os.Stderr = stderr
	got = readFile(t, "log.txt")

	if got != want {
		t.Errorf("rotate() failed\nwant: '%s'\n got: '%v'", want, got)
	}
}

func sendHupSignal(t *testing.T, pid int) {
	cmd := exec.Command("kill", "-HUP", fmt.Sprintf("%d", pid))
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n%s", err, debug.Stack())
		t.Errorf("error running kill -HUP %d\n", pid)
	}
	cmd.Wait()
	//<-time.After(60 * time.Second)
}

func TestSignal(t *testing.T) {
	defer remove(t, "log.txt", "log.txt.TestSignal")
	os.Remove("log.txt")
	os.Remove("log.txt.TestSignal")
	log.SetFlags(0)

	defer WriteTo("log.txt").Close()
	want := "signal test\n"
	log.Printf(want)

	err := os.Rename("log.txt", "log.txt.TestSignal")
	if err != nil {
		t.Fatalf("TestSignal(): %v\n", err)
	}

	got := readFile(t, "log.txt.TestSignal")
	if got != want {
		t.Fatalf("renamed log failed\nwant: '%s'\n got: '%s'", want, got)
	}
	sendHupSignal(t, os.Getpid())
	<-time.After(1 * time.Second)
	log.Println("after HUP")

	want = "after HUP\n"
	got = readFile(t, "log.txt")

	if got != want {
		t.Errorf("HUP test failed\nwant: '%s'\n got: '%s'", want, got)
	}
}

func TestCloseImmediately(t *testing.T) {
	os.Remove("log.txt")
	defer remove(t, "log.txt")

	rl := WriteTo("log.txt")
	rl.Close()

}
