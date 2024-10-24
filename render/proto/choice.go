package proto

import (
	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/tplcontainer"
)

// var _ XMLElement = (*ElementContainter)(nil)

type ElementContainter struct {
	base *XMLElementBase

	cur   string
	elems *tplcontainer.SequenceMap[string, *elementWithKey]
}

// Base implements XMLElement.
func (e *ElementContainter) Base() *XMLElementBase {
	return e.base
}

// MarshalXML implements XMLElement.
func (e *ElementContainter) MarshalXML(ns string, tag string) *etree.Element {
	panic("unimplemented")
}

// UnmarshalXML implements XMLElement.
func (e *ElementContainter) UnmarshalXML(ee *etree.Element) {
	panic("unimplemented")
}
