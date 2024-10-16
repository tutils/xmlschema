package proto

import "github.com/beevik/etree"

var _ XMLElement = (*T_CT_Doc)(nil)

type T_CT_Doc struct {
	base *XMLElementBase
	text string
}

func NewT_CT_Doc(base *XMLElementBase) *T_CT_Doc {
	if base == nil {
		base = NewXMLElementBase()
	}
	return &T_CT_Doc{
		base: base,
	}
}

// Base implements XMLElement.
func (e *T_CT_Doc) Base() *XMLElementBase {
	return e.base
}

func (e *T_CT_Doc) MarshalXML(ns string, tag string) *etree.Element {
	eb := e.Base()
	// if eb == nil {
	// 	eb = NewXMLElementBase()
	// 	e.base = eb
	// }
	tag = eb.AutoETreeTag(ns, tag)
	ee := etree.NewElement(tag)
	eb.Marshal(ee)

	ee.SetText(e.text)
	return ee
}

func (e *T_CT_Doc) UnmarshalXML(ee *etree.Element) {
	// 1. 解析属性
	// 1.1 解析命名空间定义
	eb := e.Base()
	if eb.parent == nil {
		eb.Unmarshal(ee)
	}

	// 1.2 解析其他属性值
	// 1.2.1 其他属性值初始化

	// 1.2.2 解析其他属性值

	// 2.解析元素
	// 2.1 元素初始化

	// 2.2 解析元素

	// [3.解析文本]
	e.text = ee.Text()
}

func (e *T_CT_Doc) SetText(v string) {
	e.text = v
}
