package proto

import "github.com/beevik/etree"

var _ XMLElement = (*T_CT_DefBySeq)(nil)

type T_CT_DefBySeq struct {
	base *XMLElementBase

	seqLst *SequenceList

	a_name *string
}

func NewT_CT_DefBySeq(base *XMLElementBase) *T_CT_DefBySeq {
	if base == nil {
		base = NewXMLElementBase()
	}
	restr := []*XMLElementRestriction{
		NewXMLElementRestriction("http://tutils.com", "a", 1, 1),
		NewXMLElementRestriction("http://tutils.com", "b", 0, 1),
		NewXMLElementRestriction("http://tutils.com", "c", 0, 1),
		NewXMLElementRestriction("http://tutils.com", "b", 1, 1),
		NewXMLElementRestriction("http://tutils.com", "d", 1, 1),
	}
	return &T_CT_DefBySeq{
		base:   base,
		seqLst: NewSequenceList(base, restr, 0, -1),
	}
}

// Base implements XMLElement.
func (e *T_CT_DefBySeq) Base() *XMLElementBase {
	return e.base
}

// MarshalXML implements XMLElement.
func (e *T_CT_DefBySeq) MarshalXML(ns string, tag string) *etree.Element {
	eb := e.Base()
	tag = eb.AutoETreeTag(ns, tag)
	ee := etree.NewElement(tag)
	eb.Marshal(ee)

	if e.a_name != nil {
		ee.CreateAttr("name", *e.a_name)
	}

	// define by sequence
	e.seqLst.MarshalXML(ee)

	return ee
}

// UnmarshalXML implements XMLElement.
func (e *T_CT_DefBySeq) UnmarshalXML(ee *etree.Element) {
	// 1. 解析属性
	// 1.1 解析命名空间定义
	eb := e.Base()
	if eb.parent == nil {
		eb.Unmarshal(ee)
	}

	// 1.2 解析其他属性值
	// 1.2.1 其他属性值初始化
	e.a_name = nil

	// 1.2.2 解析其他属性值
	for _, attr := range ee.Attr {
		if attr.Key == "name" && attr.Space == "" {
			e.a_name = dup(attr.Value)
			continue
		}
	}

	// 2.解析元素
	// 2.1 元素初始化
	e.seqLst = nil

	// 2.2 解析元素
	for _, child := range ee.Child {
		cee, ok := child.(*etree.Element)
		if !ok {
			continue
		}

		base := &XMLElementBase{}
		base.Unmarshal(cee)
		base.SetParent(e.base)
		// TODO:

		// if base.VarifyETreeTag(
		// 	cee,
		// 	"http://tutils.com",
		// 	"adoc",
		// ) {
		// 	ce := NewT_CT_Name(base)
		// 	ce.UnmarshalXML(cee)
		// 	e.seqLst = append(e.seqLst, NewElementWrap(
		// 		"http://tutils.com",
		// 		"adoc",
		// 		ce,
		// 	))
		// 	continue
		// }

		// if base.VarifyETreeTag(
		// 	cee,
		// 	"http://tutils.com",
		// 	"bdoc",
		// ) {
		// 	ce := NewT_CT_Name(base)
		// 	ce.UnmarshalXML(cee)
		// 	e.seqLst = append(e.seqLst, NewElementWrap(
		// 		"http://tutils.com",
		// 		"bdoc",
		// 		ce,
		// 	))
		// 	continue
		// }
	}
}

func (e *T_CT_DefBySeq) AddElemA(ce *T_CT_Doc) {
	wrap := NewElementWrap(
		"http://tutils.com",
		"a",
		ce,
	)
	e.seqLst.AddElement(wrap)
}

func (e *T_CT_DefBySeq) AddElemB(ce *T_CT_Doc) {
	wrap := NewElementWrap(
		"http://tutils.com",
		"b",
		ce,
	)
	e.seqLst.AddElement(wrap)
}

func (e *T_CT_DefBySeq) AddElemC(ce *T_CT_Doc) {
	wrap := NewElementWrap(
		"http://tutils.com",
		"c",
		ce,
	)
	e.seqLst.AddElement(wrap)
}

func (e *T_CT_DefBySeq) AddElemD(ce *T_CT_Doc) {
	wrap := NewElementWrap(
		"http://tutils.com",
		"d",
		ce,
	)
	e.seqLst.AddElement(wrap)
}

func (e *T_CT_DefBySeq) SetAttrName(v string) {
	e.a_name = &v
}
