package proto

import "github.com/beevik/etree"

var _ XMLElement = (*XMLElementRaw)(nil)

type attributeWithKey struct {
	ns    string
	tag   string
	value string
}

type elementWithKey struct {
	ns  string
	tag string
	e   XMLElement
}

type XMLElementRaw struct {
	base *XMLElementBase

	attrList []*attributeWithKey

	elemList []*elementWithKey
	text     string
}

func NewXMLElementRaw(base *XMLElementBase) *XMLElementRaw {
	return &XMLElementRaw{
		base: base,
	}
}

// Base implements XMLElement.
func (e *XMLElementRaw) Base() *XMLElementBase {
	return e.base
}

// MarshalXML implements XMLElement.
func (e *XMLElementRaw) MarshalXML(ns string, tag string) *etree.Element {
	eb := e.Base()
	tag = eb.AutoETreeTag(ns, tag)
	ee := etree.NewElement(tag)
	eb.Marshal(ee)

	for _, ak := range e.attrList {
		key := ak.tag
		if len(ak.ns) > 0 {
			// ref attr
			prefix := eb.MustGetPrefix(ak.ns)
			key = prefix + ":" + key
		}
		ee.CreateAttr(key, ak.value)
	}

	// element
	for _, ek := range e.elemList {
		cee := ek.e.MarshalXML(ek.ns, ek.tag)
		ee.AddChild(cee)
	}

	if len(e.text) > 0 {
		ee.SetText(e.text)
	}

	return ee
}

// UnmarshalXML implements XMLElement.
func (e *XMLElementRaw) UnmarshalXML(ee *etree.Element) {
	// 1. 解析属性
	// 1.1 解析命名空间定义
	eb := e.Base()
	if eb.parent == nil {
		eb.Unmarshal(ee)
	}

	// 1.2 解析其他属性值
	// 1.2.1 其他属性值初始化
	e.attrList = nil

	// 1.2.2 解析其他属性值
	for _, attr := range ee.Attr {
		if attr.Space == "xmlns" || (attr.Key == "xmlns" && attr.Space == "") {
			continue
		}

		var ns string
		if attr.Space != "" {
			ns = attr.NamespaceURI()
		}
		e.attrList = append(e.attrList, &attributeWithKey{
			ns:    ns,
			tag:   attr.Key,
			value: attr.Value,
		})
	}

	// 2.解析元素
	// 2.1 元素初始化
	e.elemList = nil

	// 2.2 解析元素
	for _, child := range ee.Child {
		cee, ok := child.(*etree.Element)
		if !ok {
			continue
		}

		base := &XMLElementBase{}
		base.Unmarshal(cee)
		base.SetParent(e.base)

		ce := NewXMLElementRaw(base)
		ce.UnmarshalXML(cee)
		e.elemList = append(e.elemList, &elementWithKey{
			ns:  cee.NamespaceURI(),
			tag: cee.Tag,
			e:   ce,
		})
	}

	// 2.3 解析文本
	e.SetText(ee.Text())
}

func (e *XMLElementRaw) AddAttribute(ns string, tag string, value string) {
	eb := e.Base()

	if ns == "xmlns" {
		eb.AddXMLNS(tag, value)
		return
	}

	if tag == "xmlns" && ns == "" {
		eb.AddXMLNS("", value)
		return
	}

	e.attrList = append(e.attrList, &attributeWithKey{
		ns:    ns,
		tag:   tag,
		value: value,
	})
}

func (e *XMLElementRaw) AddElement(ns string, tag string, ce XMLElement) {
	ce.Base().SetParent(e.Base())
	e.elemList = append(e.elemList, &elementWithKey{
		ns:  ns,
		tag: tag,
		e:   ce,
	})
}

func (e *XMLElementRaw) SetText(text string) {
	e.text = text
}

func (e *XMLElementRaw) Text() string {
	return e.text
}
