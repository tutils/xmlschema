package proto

import "github.com/beevik/etree"

var _ XMLElement = (*Choice)(nil)

type Choice struct {
	base *XMLElementBase

	attrList []*attributeWithKey

	elemList []*elementWithKey
	text     string
}

// Base implements XMLElement.
func (e *Choice) Base() *XMLElementBase {
	return e.base
}

// MarshalXML implements XMLElement.
func (e *Choice) MarshalXML(ns string, tag string) *etree.Element {
	panic("unimplemented")
}

// UnmarshalXML implements XMLElement.
func (e *Choice) UnmarshalXML(ee *etree.Element) {
	panic("unimplemented")
}
