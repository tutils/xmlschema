package proto

import "github.com/beevik/etree"

var _ XMLElement = (*T_CT_Name)(nil)

type T_CT_Name struct {
	base *XMLElementBase
	// tag  string

	a_en *string
}

func NewT_CT_Name(base *XMLElementBase) *T_CT_Name {
	if base == nil {
		base = NewXMLElementBase()
	}
	return &T_CT_Name{
		base: base,
	}
}

// GetXMLElementBase implements XMLElement.
func (e *T_CT_Name) Base() *XMLElementBase {
	return e.base
}

func (e *T_CT_Name) MarshalXML(ns string, tag string) *etree.Element {
	eb := e.Base()
	// if eb == nil {
	// 	eb = NewXMLElementBase()
	// 	e.base = eb
	// }
	tag = eb.AutoETreeTag(ns, tag)
	ee := etree.NewElement(tag)
	eb.Marshal(ee)

	if e.a_en != nil {
		ee.CreateAttr("en", *e.a_en)
	}

	return ee
}

func (e *T_CT_Name) UnmarshalXML(ee *etree.Element) {
	// 1. 解析属性
	// 1.1 解析命名空间定义
	eb := e.Base()
	if eb.parent == nil {
		eb.Unmarshal(ee)
	}

	// 1.2 解析其他属性值
	// 1.2.1 其他属性值初始化
	e.a_en = nil

	// 1.2.2 解析其他属性值
	for _, attr := range ee.Attr {
		if attr.Key == "en" && attr.Space == "" {
			e.a_en = dup(attr.Value)
			continue
		}
	}

	// 2.解析元素
	// 2.1 元素初始化

	// 2.2 解析元素
}

func (e *T_CT_Name) SetAttrEn(v string) {
	e.a_en = &v
}
