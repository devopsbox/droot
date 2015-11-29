package osutil

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	fp "path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/yuuki1/droot/log"
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

const compressionBufSize = 32768

func Compress(in io.Reader) io.ReadCloser {
	pReader, pWriter := io.Pipe()
	bufWriter := bufio.NewWriterSize(pWriter, compressionBufSize)
	compressor := gzip.NewWriter(bufWriter)

	go func() {
		_, err := io.Copy(compressor, in)
		if err == nil {
			err = compressor.Close()
		}
		if err == nil {
			err = bufWriter.Flush()
		}
		if err != nil {
			pWriter.CloseWithError(err)
		} else {
			pWriter.Close()
		}
	}()

	return pReader
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

func RObindMount(src, dest string) error {
	if err := RunCmd("mount", "-o", "remount,ro,bind", src, dest); err != nil {
		return err
	}
	return nil
}

func Mounted(mountpoint string) (bool, error) {
	mntpoint, err := os.Stat(mountpoint)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	parent, err := os.Stat(fp.Join(mountpoint, ".."))
	if err != nil {
		return false, err
	}
	mntpointSt := mntpoint.Sys().(*syscall.Stat_t)
	parentSt := parent.Sys().(*syscall.Stat_t)
	return mntpointSt.Dev != parentSt.Dev, nil
}

// Unmount will unmount the target filesystem, so long as it is mounted.
func Unmount(target string, flag int) error {
	if mounted, err := Mounted(target); err != nil || !mounted {
		return err
	}
	return ForceUnmount(target, flag)
}

// ForceUnmount will force an unmount of the target filesystem, regardless if
// it is mounted or not.
func ForceUnmount(target string, flag int) (err error) {
	// Simple retry logic for unmount
	for i := 0; i < 10; i++ {
		if err = syscall.Unmount(target, flag); err == nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
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

// Symlink, but ignore already exists file.
func Symlink(oldname, newname string) error {
	if err := os.Symlink(oldname, newname); err != nil {
		// Ignore already created symlink
		if _, ok := err.(*os.LinkError); !ok {
			return err
		}
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

// stubs

func GetMountsByRoot(rootDir string) ([]string, error) {
	return nil, fmt.Errorf("osutil: GetMountsByRoot not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func UmountRoot(rootDir string) (err error) {
	return fmt.Errorf("osutil: UmountRoot not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func LookupGroup(id string) (int, error) {
	return -1, fmt.Errorf("osutil: LookupGroup not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func SetGroup(id string) error {
	return fmt.Errorf("osutil: SetGroup not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func LookupUser(id string) error {
	return fmt.Errorf("osutil: LookupUser not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func SetUser(id string) error {
	return fmt.Errorf("osutil: SetUser not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func DropCapabilities(keepCaps map[uint]bool) error {
	return fmt.Errorf("osutil: DropCapabilities not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

