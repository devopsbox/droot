package commands

import (
	"errors"

	"github.com/codegangsta/cli"

	"github.com/yuuki/droot/deploy"
	"github.com/yuuki/droot/log"
	"github.com/yuuki/droot/mounter"
	"github.com/yuuki/droot/osutil"
)

var CommandArgRm = "--root DESTINATION_DIRECTORY"

var CommandRm = cli.Command{
	Name:   "rm",
	Usage:  "Remove directory mounted by 'run' command",
	Action: fatalOnError(doRm),
	Flags: []cli.Flag{
		cli.StringFlag{Name: "root, r", Usage: "Root directory path for chrooted"},
	},
}

func doRm(c *cli.Context) error {
	optRootDir := c.String("root")
	if optRootDir == "" {
		cli.ShowCommandHelp(c, "run")
		return errors.New("--root option required")
	}

	rootDir, err := mounter.ResolveRootDir(optRootDir)
	if err != nil {
		return err
	}

	mnt := mounter.NewMounter(rootDir)
	if err := mnt.UmountRoot(); err != nil {
		return err
	}

	if osutil.IsSymlink(optRootDir) {
		return deploy.CleanupSymlink(optRootDir)
	}

	log.Info("-->", "Removing", rootDir)
	return osutil.RunCmd("rm", "-fr", rootDir)
}
