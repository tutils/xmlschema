package render

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/beevik/etree"
)

type XMLElement interface {
	XMLElement() *etree.Element
	AddChild(XMLElement)
}

type TypeChecker func(e XMLElement) bool

type ElementChoice struct {
	etreeElem *etree.Element
	checker   TypeChecker
	seq       []any
}

func NewElementChoice(checker TypeChecker, e XMLElement) *ElementChoice {
	return &ElementChoice{
		etreeElem: e.XMLElement(),
		checker:   checker,
	}
}

func (e *ElementChoice) XMLElement() *etree.Element {
	return e.etreeElem
}

func (e *ElementChoice) AddChild(c XMLElement) {
	e.XMLElement().AddChild(c.XMLElement())
}

func (ec *ElementChoice) AddElement(e XMLElement) {
	if !ec.checker(e) {
		return
	}

	ec.seq = append(ec.seq, e)
	ec.AddChild(e)
}

type GroupShapeNonVisual struct {
	etreeElem *etree.Element
}

func NewCT_GroupShapeNonVisual() *GroupShapeNonVisual {
	e := &GroupShapeNonVisual{}
	e.etreeElem = etree.NewElement("")
	return e
}

func (e *GroupShapeNonVisual) XMLElement() *etree.Element {
	return e.etreeElem
}

func (e *GroupShapeNonVisual) AddChild(c XMLElement) {
	e.XMLElement().AddChild(c.XMLElement())
}

type Shape struct {
	etreeElem *etree.Element
}

func NewShape() *Shape {
	e := &Shape{}
	e.etreeElem = etree.NewElement("")
	return e
}

func (e *Shape) XMLElement() *etree.Element {
	return e.etreeElem
}

func (e *Shape) AddChild(c XMLElement) {
	e.XMLElement().AddChild(c.XMLElement())
}

type GroupShape struct {
	etreeElem *etree.Element
	nvGrpSpPr *GroupShapeNonVisual // 1
	choice0   *ElementChoice       // 1
}

func checkCT_GroupShapeChoice0(e XMLElement) bool {
	switch e.(type) {
	case *Shape:
		return true
	}
	return false
}

func NewGroupShape() *GroupShape {
	e := &GroupShape{}

	e.etreeElem = etree.NewElement("")

	e.choice0 = NewElementChoice(checkCT_GroupShapeChoice0, e)

	nvGrpSpPr := NewCT_GroupShapeNonVisual()
	e.SetGroupShapeNonVisual(nvGrpSpPr)

	return e
}

func (e *GroupShape) XMLElement() *etree.Element {
	return e.etreeElem
}

func (e *GroupShape) AddChild(c XMLElement) {
	e.XMLElement().AddChild(c.XMLElement())
}

func (e *GroupShape) SetGroupShapeNonVisual(c *GroupShapeNonVisual) {
	c.XMLElement().Tag = "nvGrpSpPr"
	e.nvGrpSpPr = c
	e.AddChild(c)
}

func (e *GroupShape) AddShape(c *Shape) {
	c.XMLElement().Tag = "sp"
	e.choice0.AddElement(c)
}

func TestSample() {
	e := NewGroupShape()
	e.XMLElement().Tag = "spTree"

	sp1 := NewShape()
	e.AddShape(sp1)
	sp2 := NewShape()
	e.AddShape(sp2)

	doc := etree.NewDocumentWithRoot(e.XMLElement())
	doc.Indent(4)
	doc.WriteTo(os.Stdout)
}

func TestSample2() {
	bs, _ := os.ReadFile("data/example/persion.xml")
	doc := etree.NewDocument()
	err := doc.ReadFromBytes(bs)
	root := doc.Root()
	// var doc Persion
	// err := xml.Unmarshal(bs, &doc)
	fmt.Println(err, root, "debug")
}

type Name struct {
	XMLName xml.Name
	CN      string `xml:"cn,attr"`
	EN      string `xml:"en,attr"`
}

type Persion struct {
	Sex     string `xml:"sex,attr"`
	Age     uint16 `xml:"age,attr"`
	Name    Name   `xml:"name"`
	Remark  string `xml:"remark"`
}

// MarshalXML implements xml.Marshaler.
func (p *Persion) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	panic("unimplemented")
}

// UnmarshalXML implements xml.Unmarshaler.
func (p *Persion) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {

	return nil
}

var (
	_ xml.Unmarshaler = (*Persion)(nil)
	_ xml.Marshaler   = (*Persion)(nil)
)
