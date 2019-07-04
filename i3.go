package i3

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"time"

	"github.com/pkg/errors"
)

type Block struct {
	FullText  string      `json:"full_text,omitempty"`
	ShortText string      `json:"short_text,omitempty"`
	Color     color.Color `json:"color,omitempty"`
}

type Barrer interface {
	Bar() []Block
}

type Blocker interface {
	Block() Block
}

type Status []interface{}

func (s Status) Bar() []Block {
	var objs []Block
	for _, obj := range s {
		switch o := obj.(type) {
		case Block:
			objs = append(objs, o)
		case Blocker:
			objs = append(objs, o.Block())
		case string:
			objs = append(objs, Block{FullText: o})
		case fmt.Stringer:
			objs = append(objs, Block{FullText: o.String()})
		}
	}
	return objs
}

type Runner struct {
	Frequency time.Duration
	Barrer    Barrer
}

// type out struct {
// 	Block
// 	Name     string
// 	Instance string
// }

func (r Runner) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	_, err := out.Write([]byte(`{"version": 1}
[
`))
	if err != nil {
		return errors.Wrap(err, "error writing header")
	}

	tk := time.NewTicker(r.Frequency)
	enc := json.NewEncoder(out)

	for {
		select {
		// case e := <-clicks:
		case <-tk.C:
			err = enc.Encode(r.Barrer.Bar())
			if err != nil {
				return errors.Wrap(err, "could not encode bar")
			}
			out.Write([]byte{','})
		case <-ctx.Done():
			return nil
		}
	}
}
