package i3

type Handler interface {
	Handle(c Click) bool
}

type HandlerFunc func(Click)

func (chf HandlerFunc) Handle(c Click) bool {
	chf(c)
	return false
}
