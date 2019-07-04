package i3

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Bool is a sad helper function
func Bool(v bool) *bool {
	return &v
}

// A Block represents an individual "block" (per https://i3wm.org/docs/i3bar-protocol.html)
type Block struct {
	Name                string `json:"name"`
	Instance            string `json:"instance"`
	FullText            string `json:"full_text,omitempty"`
	ShortText           string `json:"short_text,omitempty"`
	Color               *Color `json:"color,omitempty"`
	Handler             `json:"-"`
	Background          *Color `json:"background,omitempty"`
	Border              *Color `json:"border,omitempty"`
	MinWidth            int    `json:"min_width,omitempty"`
	Align               string `json:"align,omitempty"`
	Urgent              *bool  `json:"urgent,omitempty"`
	Separator           *bool  `json:"separator,omitempty"`
	SeparatorBlockWidth int    `json:"separator_block_width,omitempty"`
	Markup              string `json:"markup,omitempty"`
}

type Barrer interface {
	Bar() []Block
}

type Blocker interface {
	Block() Block
}

type BlockerFunc func() Block

func (b BlockerFunc) Block() Block {
	return b()
}

type Static []interface{}

func (s Static) Bar() []Block {
	var objs []Block
	for i, obj := range s {
		switch o := obj.(type) {
		case *Block:
			objs = append(objs, *o)
		case Block:
			objs = append(objs, o)
		case Blocker:
			objs = append(objs, o.Block())
		case string:
			objs = append(objs, Block{FullText: o})
		case fmt.Stringer:
			objs = append(objs, Block{FullText: o.String()})
		}
		if objs[i].Name == "" {
			objs[i].Name = fmt.Sprint(len(objs) - 1)
		}
		if objs[i].Instance == "" {
			objs[i].Instance = fmt.Sprint(len(objs) - 1)
		}
	}
	return objs
}

type Runner struct {
	Frequency time.Duration
	Barrer    Barrer

	updates chan struct{}
}

type Click struct {
	Name         string
	Instance     string
	X, Y, Button int
}

func (r *Runner) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	_, err := out.Write([]byte(`{"version": 1, "click_events": true}
[
`))
	if err != nil {
		return errors.Wrap(err, "error writing header")
	}

	var eg errgroup.Group

	clicks := make(chan Click)
	eg.Go(func() error {
		dec := json.NewDecoder(in)
		dec.Token()
		var click Click
		for {
			err := dec.Decode(&click)
			if err != nil {
				return errors.Wrap(err, "could not decode clicks")
			}
			clicks <- click
		}
	})

	enc := json.NewEncoder(out)
	eg.Go(func() error {
		for {
			bar := r.Barrer.Bar()
			err = enc.Encode(bar)
			if err != nil {
				return errors.Wrap(err, "could not encode bar")
			}
			out.Write([]byte{','})
			select {
			case c := <-clicks:
				if h, ok := r.Barrer.(Handler); ok && h.Handle(c) {
					continue
				}
				if n, err := strconv.Atoi(c.Instance); err == nil && bar[n].Handler != nil {
					go bar[n].Handler.Handle(c)
				}
			case <-time.After(r.Frequency):
			case <-ctx.Done():
				return nil
			}
		}
	})

	return eg.Wait()
}

type SFunc func() string

func (s SFunc) String() string {
	return s()
}

type Color struct {
	colorful.Color
}

func (c *Color) MarshalJSON() ([]byte, error) {
	return []byte("\"" + c.Hex() + "\""), nil
}
