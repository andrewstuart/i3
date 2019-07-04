package i3

import "fmt"

type Pager struct {
	Bars []Barrer

	i uint
}

func (p *Pager) Bar() []Block {
	n := p.i % uint(len(p.Bars))
	bar := p.Bars[n].Bar()
	bar = append(bar, Block{
		FullText: fmt.Sprintf("Page %02d", n+1),
	})
	return bar
}

func (p *Pager) Handle(c Click) bool {
	switch c.Button {
	case 4:
		p.i++
		return true
	case 5:
		p.i--
		return true
	}
	return false
}
