package server

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const stackFileLimit = 4 * 1024 * 1024

func openStackRoot(serverName, stackName string, create bool) (*os.Root, error) {
	if err := validateStackSegment(serverName, "server"); err != nil {
		return nil, err
	}
	if err := validateStackSegment(stackName, "stack"); err != nil {
		return nil, err
	}
	if create {
		if err := os.MkdirAll(stacksBaseDir, 0o700); err != nil {
			return nil, err
		}
	}
	base, err := os.OpenRoot(stacksBaseDir)
	if err != nil {
		return nil, err
	}
	defer base.Close()
	relative := filepath.Join(serverName, stackName)
	if create {
		if err := base.MkdirAll(relative, 0o700); err != nil {
			return nil, err
		}
	}
	return base.OpenRoot(relative)
}

func readStackFile(root *os.Root, name string) ([]byte, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("stack file must be a regular file, not a link or device")
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("stack file must be a regular file")
	}
	if err := file.Chmod(0o600); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, stackFileLimit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > stackFileLimit {
		return nil, errors.New("stack file exceeds the 4 MiB limit")
	}
	return data, nil
}

func loadStackDetail(serverName, stackName string) (stackDetailResponse, error) {
	root, err := openStackRoot(serverName, stackName, false)
	if err != nil {
		return stackDetailResponse{}, err
	}
	defer root.Close()
	return loadStackDetailFromRoot(root, stackName)
}

func loadStackDetailFromRoot(root *os.Root, stackName string) (stackDetailResponse, error) {
	if err := root.Chmod(".", 0o700); err != nil {
		return stackDetailResponse{}, err
	}
	content, err := readStackFile(root, "docker-compose.yml")
	if err != nil {
		return stackDetailResponse{}, err
	}
	detail := stackDetailResponse{Name: stackName, ComposeYml: string(content)}
	env, err := readStackFile(root, ".env")
	if err == nil {
		detail.Env = string(env)
		detail.HasEnv = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		return stackDetailResponse{}, err
	}
	if info, err := root.Stat("docker-compose.yml"); err == nil {
		detail.UpdatedAt = info.ModTime().Format(time.RFC3339)
	}
	return detail, nil
}

func saveStackFiles(serverName, stackName, composeYml, env string, useEnv bool) error {
	if len(composeYml) > stackFileLimit || (useEnv && len(env) > stackFileLimit) {
		return errors.New("stack file exceeds the 4 MiB limit")
	}
	root, err := openStackRoot(serverName, stackName, true)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Chmod(".", 0o700); err != nil {
		return err
	}
	if err := writePrivateFileInRoot(root, "docker-compose.yml", []byte(composeYml)); err != nil {
		return err
	}
	if useEnv {
		return writePrivateFileInRoot(root, ".env", []byte(env))
	}
	if err := root.Remove(".env"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func writePrivateFileAtomically(path string, data []byte) error {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer root.Close()
	return writePrivateFileInRoot(root, filepath.Base(path), data)
}

func writePrivateFileInRoot(root *os.Root, name string, data []byte) error {
	if len(data) > stackFileLimit {
		return errors.New("stack file exceeds the 4 MiB limit")
	}
	token, err := generateToken(16)
	if err != nil {
		return err
	}
	temporary := ".contiwatch-stack-" + token + ".tmp"
	file, err := root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
		_ = root.Remove(temporary)
	}()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return root.Rename(temporary, name)
}
