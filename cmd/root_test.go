package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--help"})

	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing root command: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("exifrenamer [source_paths...] [flags]")) {
		t.Errorf("expected help output to contain usage text, got: %s", output)
	}
}

func TestRootCmd_NoArgs(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{})

	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing root command with no args: %v", err)
	}
}
