package commands

import (
	"io"

	cmds "github.com/jbenet/go-ipfs/commands"
	core "github.com/jbenet/go-ipfs/core"
	ccutil "github.com/jbenet/go-ipfs/core/commands/util"
	path "github.com/jbenet/go-ipfs/path"
	uio "github.com/jbenet/go-ipfs/unixfs/io"

	"github.com/jbenet/go-ipfs/Godeps/_workspace/src/github.com/cheggaaa/pb"
)

var CatCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Show IPFS object data",
		ShortDescription: `
Retrieves the object named by <ipfs-path> and outputs the data
it contains.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ipfs-path", true, true, "The path to the IPFS object(s) to be outputted").EnableStdin(),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		node, err := req.Context().GetNode()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		readers := make([]io.Reader, 0, len(req.Arguments()))

		readers, length, err := cat(node, req.Arguments())
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		res.SetLength(length)

		reader := io.MultiReader(readers...)
		res.SetOutput(reader)
	},
	PostRun: func(req cmds.Request, res cmds.Response) {
		if res.Length() < ccutil.ProgressBarMinSize {
			return
		}

		bar := pb.New(int(res.Length())).SetUnits(pb.U_BYTES)
		bar.Output = res.Stderr()
		bar.Start()

		reader := bar.NewProxyReader(res.Output().(io.Reader))
		res.SetOutput(reader)
	},
}

func cat(node *core.IpfsNode, paths []string) ([]io.Reader, uint64, error) {
	readers := make([]io.Reader, 0, len(paths))
	length := uint64(0)
	for _, fpath := range paths {
		dagnode, err := node.Resolver.ResolvePath(path.Path(fpath))
		if err != nil {
			return nil, 0, err
		}

		read, err := uio.NewDagReader(node.Context(), dagnode, node.DAG)
		if err != nil {
			return nil, 0, err
		}
		readers = append(readers, read)
		length += uint64(read.Size())
	}
	return readers, length, nil
}
