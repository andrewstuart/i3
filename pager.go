package i3

type Pager struct {
	Bars []Barrer

	i uint
}

func (p *Pager) Bar() []Block {
	return p.Bars[p.i%uint(len(p.Bars))].Bar()
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
