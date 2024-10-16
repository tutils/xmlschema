package proto

import "github.com/beevik/etree"

var _ XMLElement = (*T_CT_Person)(nil)

type T_CT_Person struct {
	base *XMLElementBase
	// tag  string

	e_t_name     *T_CT_Name
	e_t_remark   *T_CT_Doc
	e_t_arrName  []*T_CT_Name
	e_t_defbyseq *T_CT_DefBySeq

	// AG_PersonBase
	a_sex *string
	a_age *string

	// ref
	a_t_url *string
}

func NewT_CT_Person(base *XMLElementBase) *T_CT_Person {
	if base == nil {
		base = NewXMLElementBase()
	}
	return &T_CT_Person{
		base: base,
	}
}

// Base implements XMLElement.
func (e *T_CT_Person) Base() *XMLElementBase {
	return e.base
}

func (e *T_CT_Person) MarshalXML(ns string, tag string) *etree.Element {
	eb := e.Base()
	tag = eb.AutoETreeTag(ns, tag)
	ee := etree.NewElement(tag)
	eb.Marshal(ee)

	if e.a_sex != nil {
		ee.CreateAttr("sex", *e.a_sex)
	}

	if e.a_age != nil {
		ee.CreateAttr("age", *e.a_age)
	}

	// ref attr
	if e.a_t_url != nil {
		prefix := eb.MustGetPrefix("http://tutils.com")
		// TODO: 后续去掉这段断言
		if prefix == "" {
			panic("invalid ref attr")
		}
		ee.CreateAttr(prefix+":url", *e.a_t_url)
	}

	// element
	if ce := e.e_t_name; ce != nil {
		cee := ce.MarshalXML("http://tutils.com", "name")
		ee.AddChild(cee)
	}

	if ce := e.e_t_remark; ce != nil {
		cee := ce.MarshalXML("http://tutils.com", "remark")
		ee.AddChild(cee)
	}

	for _, ce := range e.e_t_arrName {
		cee := ce.MarshalXML("http://tutils.com", "arrName")
		ee.AddChild(cee)
	}

	if ce := e.e_t_defbyseq; ce != nil {
		cee := ce.MarshalXML("http://tutils.com", "defbyseq")
		ee.AddChild(cee)
	}

	return ee
}

func (e *T_CT_Person) UnmarshalXML(ee *etree.Element) {
	// 1. 解析属性
	// 1.1 解析命名空间定义
	eb := e.Base()
	if eb.parent == nil {
		eb.Unmarshal(ee)
	}

	// 1.2 解析其他属性值
	// 1.2.1 其他属性值初始化
	e.a_sex = nil
	e.a_age = nil
	e.a_t_url = nil

	// 1.2.2 解析其他属性值
	for _, attr := range ee.Attr {
		if attr.Key == "sex" && attr.Space == "" {
			e.a_sex = dup(attr.Value)
			continue
		}

		if attr.Key == "age" && attr.Space == "" {
			e.a_age = dup(attr.Value)
			continue
		}

		// ref attribute
		if attr.Key == "url" && attr.Space == eb.MustGetPrefix("http://tutils.com") {
			e.a_t_url = dup(attr.Value)
			continue
		}
	}

	// 2.解析元素
	// 2.1 元素初始化
	e.e_t_name = nil
	e.e_t_remark = nil

	// 2.2 解析元素
	for _, child := range ee.Child {
		cee, ok := child.(*etree.Element)
		if !ok {
			continue
		}

		base := &XMLElementBase{}
		base.Unmarshal(cee)
		base.SetParent(e.base)

		if base.VarifyETreeTag(cee, "http://tutils.com", "name") {
			ce := NewT_CT_Name(base)
			ce.UnmarshalXML(cee)
			e.e_t_name = ce
			continue
		}

		if base.VarifyETreeTag(cee, "http://tutils.com", "remark") {
			ce := NewT_CT_Doc(base)
			ce.UnmarshalXML(cee)
			e.e_t_remark = ce
			continue
		}

		if base.VarifyETreeTag(cee, "http://tutils.com", "arrName") {
			ce := NewT_CT_Name(base)
			ce.UnmarshalXML(cee)
			e.e_t_arrName = append(e.e_t_arrName, ce)
			continue
		}

		if base.VarifyETreeTag(cee, "http://tutils.com", "defbyseq") {
			ce := NewT_CT_DefBySeq(base)
			ce.UnmarshalXML(cee)
			e.e_t_defbyseq = ce
			continue
		}
	}
}

func (e *T_CT_Person) SetElemName(ce *T_CT_Name) {
	ce.Base().SetParent(e.Base())
	e.e_t_name = ce
}

func (e *T_CT_Person) SetElemRemark(ce *T_CT_Doc) {
	ce.Base().SetParent(e.Base())
	e.e_t_remark = ce
}

func (e *T_CT_Person) AddElemArrName(ce *T_CT_Name) {
	ce.Base().SetParent(e.Base())
	e.e_t_arrName = append(e.e_t_arrName, ce)
}

func (e *T_CT_Person) SetElemDefbyseq(ce *T_CT_DefBySeq) {
	ce.Base().SetParent(e.Base())
	e.e_t_defbyseq = ce
}

func (e *T_CT_Person) SetAttrSex(v string) {
	e.a_sex = &v
}

func (e *T_CT_Person) SetAttrAge(v string) {
	e.a_age = &v
}

func (e *T_CT_Person) SetAttrTURL(v string) {
	e.a_t_url = &v
}
