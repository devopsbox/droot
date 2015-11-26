package osutil

import (
	"compress/gzip"
	"errors"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/yuuki1/dochroot/log"
)

var RsyncDefaultOpts = []string{"-av", "--delete"}

func ExistsFile(file string) bool {
	f, err := os.Stat(file)
	return err == nil && !f.IsDir()
}

func ExistsDir(dir string) bool {
	if f, err := os.Stat(dir); os.IsNotExist(err) || !f.IsDir() {
		return false
	}
	return true
}

func IsDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func RunCmd(name string, arg ...string) error {
	out, err := exec.Command(name, arg...).CombinedOutput()
	if len(out) > 0 {
		log.Debug(string(out))
	}
	if err != nil {
		log.Errorf("failed: %s %s", name, arg)
		return err
	}
	log.Debug("runcmd: ", name, arg)
	return nil
}

func Gzip(destWriter io.Writer, srcReader io.Reader) error {
	w := gzip.NewWriter(destWriter)
	defer w.Close()

	bytes, err := ioutil.ReadAll(srcReader)
	if err != nil {
		return err
	}

	nBytes, err := w.Write(bytes)
	if err != nil {
		return err
	}
	log.Debug("gzip bytes", nBytes)

	return nil
}

func ExtractTarGz(filePath string) error {
	if err := RunCmd("tar", "xf", filePath); err != nil {
		return err
	}

	if err := os.Chmod(filePath, os.FileMode(0755)); err != nil {
		return err
	}

	return nil
}

func Rsync(from, to string, arg ...string) error {
	from = from + "/"
	// append "/" when not terminated by "/"
	if strings.LastIndex(to, "/") != len(to)-1 {
		to = to + "/"
	}

	// TODO --exclude, --excluded-from
	rsyncArgs := []string{}
	rsyncArgs = append(rsyncArgs, RsyncDefaultOpts...)
	rsyncArgs = append(rsyncArgs, from, to)
	if err := RunCmd("rsync", rsyncArgs...); err != nil {
		return err
	}

	return nil
}

func Cp(from, to string) error {
	if err := RunCmd("cp", "-p", from, to); err != nil {
		return err
	}
	return nil
}

func BindMount(src, dest string) error {
	if err := RunCmd("mount", "--bind", src, dest); err != nil {
		return err
	}
	return nil
}

func DropCapabilities(keepCaps map[uint]bool) error {
	var i uint
	for i = 0; ; i++ {
		if keepCaps[i] {
			continue
		}
		if err := unix.Prctl(syscall.PR_CAPBSET_READ, uintptr(i), 0, 0, 0); err != nil {
			// Regard EINVAL as the condition of loop finish.
			if errno, ok := err.(syscall.Errno); ok && errno == syscall.EINVAL {
				break
			}
			return err
		}
		if err := unix.Prctl(syscall.PR_CAPBSET_DROP, uintptr(i), 0, 0, 0); err != nil {
			// Ignore EINVAL since the capability may not be supported in this system.
			if errno, ok := err.(syscall.Errno); ok && errno == syscall.EINVAL {
				continue
			} else if errno, ok := err.(syscall.Errno); ok && errno == syscall.EPERM {
				return errors.New("required CAP_SETPCAP capabilities")
			} else {
				return err
			}
		}
	}

	if i == 0 {
		return errors.New("Failed to drop capabilities")
	}

	return nil
}

// Mknod unless path does not exists.
func Mknod(path string, mode uint32, dev int) error {
	if ExistsFile(path) {
		return nil
	}
	if err := syscall.Mknod(path, mode, dev); err != nil {
		return err
	}
	return nil
}

func Execv(cmd string, args []string, env []string) error {
	name, err := exec.LookPath(cmd)
	if err != nil {
		return err
	}

	log.Debug("exec: ", name, args)

	return syscall.Exec(name, args, env)
}

func ChrootAndExec(keepCaps map[uint]bool, rootDir string, command ...string) error {
	if err := syscall.Chroot(rootDir); err != nil {
		return err
	}
	if err := syscall.Chdir("/"); err != nil {
		return err
	}
	if err := DropCapabilities(keepCaps); err != nil {
		return err
	}
	return Execv(command[0], command[0:], os.Environ())
}
