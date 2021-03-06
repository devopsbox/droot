package osutil

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistsFile(t *testing.T) {
	assert.False(t, ExistsFile("/paht/to/notexist"))

	tmpDir := os.TempDir()
	tmp, _ := ioutil.TempFile(tmpDir, "droot_test")
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	assert.True(t, ExistsFile(tmp.Name()))
	assert.False(t, ExistsFile(tmpDir))
}

func TestIsSymlink(t *testing.T) {
	tmpDir := os.TempDir()
	tmp, _ := ioutil.TempFile(tmpDir, "droot_test")
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	assert.False(t, IsSymlink(tmp.Name()))

	os.Symlink(tmp.Name(), tmpDir+"/symlink")

	assert.True(t, IsSymlink(tmpDir+"/symlink"))
}

func TestExistsDir(t *testing.T) {
	assert.False(t, ExistsDir("/paht/to/notexist"))

	tmpDir := os.TempDir()
	tmp, _ := ioutil.TempFile(tmpDir, "droot_test")
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	assert.True(t, ExistsDir(tmpDir))
	assert.False(t, ExistsDir(tmp.Name()))
}

func TestIsDirEmpty(t *testing.T) {
	ok := IsDirEmpty("/paht/to/notexist")
	assert.False(t, ok)

	tmpDir := os.TempDir()
	os.Mkdir(tmpDir+"/empty", 0755)
	os.Mkdir(tmpDir+"/noempty", 0755)
	os.Create(tmpDir + "/noempty/test")
	defer func() {
		os.Remove(tmpDir + "/empty")
		os.RemoveAll(tmpDir + "/noempty")
	}()

	ok = IsDirEmpty(tmpDir + "/empty")
	assert.True(t, ok)

	ok = IsDirEmpty(tmpDir + "/noempty")
	assert.False(t, ok)
}

func TestRunCmd(t *testing.T) {
	assert.NoError(t, RunCmd("/bin/ls"))
	assert.Error(t, RunCmd("/bin/hoge"))
}

func TestSymlink(t *testing.T) {
	tmpDir := os.TempDir()
	tmp, _ := ioutil.TempFile(tmpDir, "droot_test")
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	assert.NoError(t, Symlink(tmp.Name(), tmp.Name()+"/symlink"))
	assert.NoError(t, Symlink(tmp.Name(), tmp.Name()+"/symlink"), "Ignore already exist symlink file")
	os.Create(tmpDir + "/droot_dummy")
	assert.NoError(t, Symlink(tmp.Name(), tmpDir+"/droot_dummy"), "Ignore already exist file")
	os.Remove(tmp.Name() + "/symlink")
}
