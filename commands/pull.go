package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/codegangsta/cli"

	"github.com/yuuki1/droot/archive"
	"github.com/yuuki1/droot/aws"
	"github.com/yuuki1/droot/log"
	"github.com/yuuki1/droot/osutil"
)

var CommandArgPull = "--dest DESTINATION_DIRECTORY --src S3_ENDPOINT [--user USER] [--grpup GROUP]"
var CommandPull = cli.Command{
	Name:   "pull",
	Usage:  "Pull an extracted docker image from s3",
	Action: fatalOnError(doPull),
	Flags: []cli.Flag{
		cli.StringFlag{Name: "dest, d", Usage: "Local filesystem path (ex. /var/containers/app)"},
		cli.StringFlag{Name: "src, s", Usage: "Amazon S3 endpoint (ex. s3://drootexample/app.tar.gz)"},
		cli.StringFlag{Name: "user, u", Usage: "User (ID or name) to set after extracting archive (required superuser)"},
		cli.StringFlag{Name: "group, g", Usage: "Group (ID or name) to set after extracting archive (required superuser)"},
	},
}

func doPull(c *cli.Context) error {
	destDir := c.String("dest")
	srcURL := c.String("src")
	if destDir == "" || srcURL == "" {
		cli.ShowCommandHelp(c, "pull")
		return errors.New("--src and --dest option required ")
	}

	s3URL, err := url.Parse(srcURL)
	if err != nil {
		return err
	}
	if s3URL.Scheme != "s3" {
		return fmt.Errorf("Not s3 scheme %s", srcURL)
	}

	uid, gid := os.Getuid(), os.Getgid()
	if group := c.String("group"); group != "" {
		if gid, err = osutil.LookupGroup(group); err != nil {
			return fmt.Errorf("Failed to lookup group:", err)
		}
	}
	if user := c.String("user"); user != "" {
		if uid, err = osutil.LookupUser(user); err != nil {
			return fmt.Errorf("Failed to lookup user:", err)
		}
	}

	downloadSize, imageReader, err := aws.NewS3Client().Download(s3URL)
	if err != nil {
		return fmt.Errorf("Failed to create temporary file: %s", err)
	}
	defer imageReader.Close()
	log.Info("downloaded", "from", s3URL, downloadSize, "bytes")

	dir, err := ioutil.TempDir(os.TempDir(), "droot")
	if err != nil {
		return fmt.Errorf("Failed to download file(%s) from s3: %s", srcURL, err)
	}
	defer os.RemoveAll(dir)

	if err := archive.ExtractTarGz(imageReader, dir, uid, gid); err != nil {
		return fmt.Errorf("Failed to rsync: %s", err)
	}

	log.Info("rsync:", "from", dir, "to", destDir)
	if err := archive.Rsync(dir, destDir); err != nil {
		return fmt.Errorf("Failed to rsync: %s", err)
	}
	if err := os.Lchown(destDir, uid, gid); err != nil {
		return fmt.Errorf("Failed to chown %d:%d: %s", uid, gid, err)
	}
	if err := os.Chmod(destDir, os.FileMode(0755)); err != nil {
		return fmt.Errorf("Failed to chmod %s: %s", destDir, err)
	}

	return nil
}
