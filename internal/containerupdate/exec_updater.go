package containerupdate

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/tilt-dev/tilt/internal/build"
	"github.com/tilt-dev/tilt/internal/k8s"
	"github.com/tilt-dev/tilt/internal/store"
	"github.com/tilt-dev/tilt/pkg/logger"
	"github.com/tilt-dev/tilt/pkg/model"
)

type ExecUpdater struct {
	kCli k8s.Client
}

var _ ContainerUpdater = &ExecUpdater{}

func NewExecUpdater(kCli k8s.Client) *ExecUpdater {
	return &ExecUpdater{kCli: kCli}
}

func (cu *ExecUpdater) UpdateContainer(ctx context.Context, cInfo store.ContainerInfo,
	archiveToCopy io.Reader, filesToDelete []string, cmds []model.Cmd, hotReload bool) error {
	if !hotReload {
		return fmt.Errorf("ExecUpdater does not support `restart_container()` step. If you ran Tilt " +
			"with `--updateMode=exec`, omit this flag. If you are using a non-Docker container runtime, " +
			"see https://github.com/tilt-dev/rerun-process-wrapper for a workaround")
	}

	l := logger.Get(ctx)
	w := logger.Get(ctx).Writer(logger.InfoLvl)

	// delete files (if any)
	if len(filesToDelete) > 0 {
		err := cu.kCli.Exec(ctx,
			cInfo.PodID, cInfo.ContainerName, cInfo.Namespace,
			append([]string{"rm", "-rf"}, filesToDelete...), nil, w, w)
		if err != nil {
			return err
		}
	}

	// copy files to container
	err := cu.kCli.Exec(ctx, cInfo.PodID, cInfo.ContainerName, cInfo.Namespace,
		[]string{"tar", "-C", "/", "-x", "-f", "-"}, archiveToCopy, w, w)
	if err != nil {
		return err
	}

	// run commands
	for i, c := range cmds {
		l.Infof("[CMD %d/%d] %s", i+1, len(cmds), strings.Join(c.Argv, " "))
		err := cu.kCli.Exec(ctx, cInfo.PodID, cInfo.ContainerName, cInfo.Namespace,
			c.Argv, nil, w, w)
		if err != nil {
			return build.WrapCodeExitError(err, cInfo.ContainerID, c)
		}

	}

	return nil
}
