package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type argsCmd struct {
	webPort int
	clear   bool
	post    httpPoster
}

func newArgsCmd() *argsCmd {
	return &argsCmd{post: http.Post}
}

func (c *argsCmd) register() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "args [<tiltfile args>]",
		Short: "Changes the Tiltfile args in use by a running Tilt",
		Long: `Changes the Tiltfile args in use by a running Tilt.

Note that this does not affect built-in Tilt args (e.g. --hud, --host), but rather the extra args that come after,
i.e., those specifying which resources to run and/or handled by a Tiltfile calling config.parse.

To provide args starting with --, insert a standalone --, e.g.:

tilt args -- --foo=bar frontend backend
`,
	}

	cmd.Flags().IntVar(&c.webPort, "port", DefaultWebPort, "Web port for the Tilt whose args should change")
	cmd.Flags().BoolVar(&c.clear, "clear", false, "Clear the Tiltfile args, as if you'd run tilt with no args")

	return cmd
}

type httpPoster func(url string, contentType string, body io.Reader) (*http.Response, error)

func (c *argsCmd) run(ctx context.Context, args []string) error {
	// require --clear instead of an empty args list to ensure an experimental `tilt flags` doesn't unintentionally wipe state
	if len(args) == 0 {
		if !c.clear {
			return errors.New("no args specified. If your intent is to empty the args, run `tilt args --clear`.")
		}
	} else {
		if c.clear {
			return errors.New("--clear cannot be specified with other values. either use --clear to clear the args or specify args to replace the args with a new (non-empty) value")
		}
	}
	url := fmt.Sprintf("http://localhost:%d/api/set_tiltfile_args", c.webPort)
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(args)
	if err != nil {
		return errors.Wrap(err, "failed to encode args as json")
	}
	res, err := c.post(url, "application/json", body)
	if err != nil {
		return errors.Wrapf(err, "could not connect to Tilt at %s", url)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error connecting to Tilt at %s: %d", url, res.StatusCode)
	}

	fmt.Printf("changed config args for Tilt running on port %d to %v\n", webPort, args)

	return nil
}
