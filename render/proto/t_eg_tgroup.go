package proto

import "github.com/beevik/etree"

var _ XMLElement = (*T_CT_Name)(nil)

type T_EG_TGroup struct {
	base *XMLElementBase
	// tag  string

	e_t_a *T_CT_Doc
	e_t_b *T_CT_Doc
	e_t_c *T_CT_Doc
}

func NewT_EG_TGroup(base *XMLElementBase) *T_EG_TGroup {
	if base == nil {
		panic("empty base")
	}
	return &T_EG_TGroup{
		base: base,
	}
}

// GetXMLElementBase implements XMLElement.
func (e *T_EG_TGroup) Base() *XMLElementBase {
	return e.base
}

func (e *T_EG_TGroup) MarshalXML(ee *etree.Element) {
	if ce := e.e_t_a; ce != nil {
		cee := ce.MarshalXML("http://tutils.com", "a")
		ee.AddChild(cee)
	}

	if ce := e.e_t_b; ce != nil {
		cee := ce.MarshalXML("http://tutils.com", "b")
		ee.AddChild(cee)
	}

	if ce := e.e_t_c; ce != nil {
		cee := ce.MarshalXML("http://tutils.com", "c")
		ee.AddChild(cee)
	}
}

func (e *T_EG_TGroup) UnmarshalXML(ee *etree.Element) {
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
	e.e_t_a = nil
	e.e_t_b = nil
	e.e_t_c = nil

	// 2.2 解析元素
	for _, child := range ee.Child {
		cee, ok := child.(*etree.Element)
		if !ok {
			continue
		}

		base := &XMLElementBase{}
		base.Unmarshal(cee)
		base.SetParent(e.base)

		if base.VarifyETreeTag(cee, "http://tutils.com", "a") {
			ce := NewT_CT_Doc(base)
			ce.UnmarshalXML(cee)
			e.e_t_a = ce
			continue
		}

		if base.VarifyETreeTag(cee, "http://tutils.com", "b") {
			ce := NewT_CT_Doc(base)
			ce.UnmarshalXML(cee)
			e.e_t_b = ce
			continue
		}

		if base.VarifyETreeTag(cee, "http://tutils.com", "c") {
			ce := NewT_CT_Doc(base)
			ce.UnmarshalXML(cee)
			e.e_t_c = ce
			continue
		}
	}
}
