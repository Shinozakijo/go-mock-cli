package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func openEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		cmd := exec.Command(editor, filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// fallback ตาม OS
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("notepad", filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case "darwin":
		// ถ้าไม่มี EDITOR ใช้ nano ก่อน
		cmd := exec.Command("nano", filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	default:
		cmd := exec.Command("nano", filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

func writeTempFile(prefix string, content []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", prefix+"-*.json")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(content); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	return tmpFile.Name(), nil
}