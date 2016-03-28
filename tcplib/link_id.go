package tcplib

type LinkID int32

type LinkIDGen struct {
	incId LinkID
}

func NewLinkIDGen() *LinkIDGen {
	return &LinkIDGen{incId: 0}
}

//start id is 1.
func (g *LinkIDGen) NewID() LinkID {
	g.incId++
	return g.incId
}

func (g *LinkIDGen) Reset() {
	g.incId = 0
}
