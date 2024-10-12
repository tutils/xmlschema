package proto

import (
	"fmt"
	"os"

	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/data"
	"github.com/tutils/xmlschema/tplcontainer"
)

/*
1. 成员前缀解决跨空间引用group和attributeGroup时，展开token后名字相同问题
2. xmlns成员，前缀与命名空间的映射

元素，引用和非引用，前缀规则一致。

属性则遵循以下规则：
1. 非引用属性，必须没有前缀
2. 引用属性，必须有前缀，哪怕引用的为当前空间的属性

组、属性组，对于引用的和非引用的，先将其展开，规则与单一元素和单一属性一致。
1. 对于属性组的引用，完全等价于直接展开；无论是本空间的还是其他空间的，均视作本空间的属性
2. 如果属性组的引用，属性组里包含引用属性，则直接视为普通引用属性
*/

// Unmarshal 步骤
// 1. 解析属性
// 1.1 解析命名空间定义

// 1.2 解析其他属性值
// 1.2.1 其他属性值初始化

// 1.2.2 解析其他属性值
// 2.解析元素
// 2.1 元素初始化
// 2.2 解析元素
// [3.解析文本]

type Token struct {
	Namespace string
	Tag       string
}

type XMLNSMap struct {
	defNS     string
	prefix2Ns *tplcontainer.SequenceMap[string, string]
	ns2prefix map[string]string
}

func NewXMLNSMap() *XMLNSMap {
	return &XMLNSMap{
		prefix2Ns: tplcontainer.NewSequenceMap[string, string](),
		ns2prefix: make(map[string]string),
	}
}

func (m *XMLNSMap) PrefixOrder() []string {
	return m.prefix2Ns.Order()
}

func (m *XMLNSMap) Set(prefix string, ns string) {
	if prefix == "" {
		panic("invalid prefix")
	}

	m.prefix2Ns.Set(prefix, ns)
	m.ns2prefix[ns] = prefix
}

func (m *XMLNSMap) GetNS(prefix string) (string, bool) {
	return m.prefix2Ns.Get(prefix)
}

func (m *XMLNSMap) GetPrefix(ns string) (string, bool) {
	prefix, ok := m.ns2prefix[ns]
	return prefix, ok
}

func (m *XMLNSMap) MustGetNS(prefix string) string {
	return m.prefix2Ns.MustGet(prefix)
}

func (m *XMLNSMap) MustGetPrefix(ns string) string {
	if prefix, ok := m.ns2prefix[ns]; ok {
		return prefix
	}
	panic("invalid key")
}

func (m *XMLNSMap) SetDefaultNS(ns string) {
	m.defNS = ns
}

func (m *XMLNSMap) DefaultNS() string {
	return m.defNS
}

// type E_CT_Work struct {
// 	xmlns XMLNSMap

// 	e_e_remark string

// 	// AG_WhereAndDuration
// 	a_name     string
// 	a_duration uint16
// }

// func (e *E_CT_Work) SetRemark(v string) {
// }

// type E_CT_Educate struct {
// 	xmlns XMLNSMap

// 	e_e_remark string
// 	e_t_remark string

// 	// AG_WhereAndDuration
// 	a_name     string
// 	a_duration uint16

// 	a_finished bool
// }

// type E_ST_EMail struct {
// }

// type E_CT_CV struct {
// 	xmlns XMLNSMap

// 	e_e_baseInfo T_CT_Person
// 	e_e_email    E_ST_EMail
// 	choice0_list any // E_CT_Educate|E_CT_Work

// 	// ref
// 	a_t_url string
// }

// type T_ST_Sex struct {
// }

// type T_ST_Age struct {
// }

func dup[T any](v T) *T {
	return &v
}

func dupPtr[T any](ptr *T) *T {
	t := *ptr
	return &t
}

// Go不支持基类方法，使用全局函数

// func MustGetPrefix(e XMLElementBase, ns string) string {
// 	for cur := e; cur != nil; cur = cur.Parent() {
// 		xmlns := e.XMLNS()
// 		if xmlns == nil {
// 			continue
// 		}

// 		// 存在命名空间定义
// 		if prefix, ok := xmlns.GetPrefix(ns); ok {
// 			return prefix
// 		}
// 	}

// 	panic("invalid ns")
// }

// func GetDefaultNS(e XMLElementBase) string {
// 	for cur := e; cur != nil; cur = cur.Parent() {
// 		xmlns := e.XMLNS()
// 		if xmlns == nil {
// 			continue
// 		}

// 		// 存在命名空间定义
// 		if defNS := xmlns.GetDefaultNS(); defNS != "" {
// 			return defNS
// 		}
// 	}

// 	return ""
// }

// func VarifyElementSpace(e *XMLElement, ee *etree.Element, ns string) bool {
// }

// func AddXMLNS(e XMLElementBase, prefix string, ns string) {
// 	xmlns := e.XMLNS()
// 	if xmlns == nil {
// 		xmlns = e.AllocXMLNS()
// 	}

// 	if prefix == "" {
// 		xmlns.SetDefaultNS(ns)
// 		return
// 	}

// 	xmlns.Set(prefix, ns)
// }

// var _ XMLElementBase = (*XMLElementBaseOnly)(nil)

type XMLElementBase struct {
	parent *XMLElementBase
	xmlns  *XMLNSMap
}

func NewXMLElementBase() *XMLElementBase {
	return &XMLElementBase{}
}

// AllocXMLNS implements XMLElementBase.
func (eb *XMLElementBase) AllocXMLNS() *XMLNSMap {
	eb.xmlns = NewXMLNSMap()
	return eb.xmlns
}

// Parent implements XMLElementBase.
func (eb *XMLElementBase) Parent() *XMLElementBase {
	return eb.parent
}

// SetParent implements XMLElementBase.
func (eb *XMLElementBase) SetParent(p *XMLElementBase) {
	eb.parent = p
}

// XMLNS implements XMLElementBase.
func (eb *XMLElementBase) XMLNS() *XMLNSMap {
	return eb.xmlns
}

func (eb *XMLElementBase) GetPrefix(ns string) (string, bool) {
	for cur := eb; cur != nil; cur = cur.Parent() {
		xmlns := cur.XMLNS()
		if xmlns == nil {
			continue
		}

		// 存在命名空间定义
		if prefix, ok := xmlns.GetPrefix(ns); ok {
			return prefix, true
		}
	}

	return "", false
}

func (eb *XMLElementBase) MustGetPrefix(ns string) string {
	prefix, ok := eb.GetPrefix(ns)
	if ok {
		return prefix
	}

	panic("invalid ns")
}

func (eb *XMLElementBase) DefaultNS() string {
	for cur := eb; cur != nil; cur = cur.Parent() {
		xmlns := eb.XMLNS()
		if xmlns == nil {
			continue
		}

		// 存在命名空间定义
		if defNS := xmlns.DefaultNS(); defNS != "" {
			return defNS
		}
	}

	return ""
}

func (eb *XMLElementBase) VarifyETreeTag(ee *etree.Element, ns string, tag string) bool {
	// 名字匹配
	if ee.Tag != tag {
		return false
	}

	// 命名空间匹配
	if ee.Space == "" {
		return eb.DefaultNS() == ns
	}
	return eb.MustGetPrefix(ns) == ee.Space
}

func (eb *XMLElementBase) AutoETreeTag(ns string, tag string) string {
	// 优先前缀
	prefix, ok := eb.GetPrefix(ns)
	if ok {
		return prefix + ":" + tag
	}

	// 其次无前缀
	if eb.DefaultNS() == ns {
		return tag
	}

	panic("invalid tag")
}

func (eb *XMLElementBase) AddXMLNS(prefix string, ns string) {
	xmlns := eb.XMLNS()
	if xmlns == nil {
		xmlns = eb.AllocXMLNS()
	}

	if prefix == "" {
		xmlns.SetDefaultNS(ns)
		return
	}

	xmlns.Set(prefix, ns)
}

func (eb *XMLElementBase) Marshal(ee *etree.Element) {
	xmlns := eb.XMLNS()
	if xmlns == nil {
		return
	}

	// 存在局部命名空间
	if defNS := xmlns.DefaultNS(); defNS != "" {
		ee.CreateAttr("xmlns", defNS)
	}

	for _, prefix := range xmlns.PrefixOrder() {
		ns := xmlns.MustGetNS(prefix)
		key := "xmlns:" + prefix
		ee.CreateAttr(key, ns)
	}
}

func (eb *XMLElementBase) Unmarshal(ee *etree.Element) {
	eb.parent = nil
	eb.xmlns = nil
	for _, attr := range ee.Attr {
		if attr.Space == "xmlns" {
			eb.AddXMLNS(attr.Key, attr.Value)
			continue
		}

		if attr.Key == "xmlns" && attr.Space == "" {
			eb.AddXMLNS("", attr.Value)
			continue
		}
	}
}

type XMLElement interface {
	Base() *XMLElementBase

	MarshalXML(ns string, tag string) *etree.Element
	UnmarshalXML(ee *etree.Element)
}

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

var _ XMLElement = (*T_CT_Person)(nil)

type T_CT_Person struct {
	base *XMLElementBase
	// tag  string

	e_t_name   *T_CT_Name
	e_t_remark *T_CT_Doc

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
	// if eb == nil {
	// 	eb = NewXMLElementBase()
	// 	e.base = eb
	// }
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
	}
}

func (e *T_CT_Person) SetElemName(ce *T_CT_Name) {
	ce.Base().SetParent(e.base)
	e.e_t_name = ce
}

func (e *T_CT_Person) SetElemRemark(ce *T_CT_Doc) {
	ce.Base().SetParent(e.base)
	e.e_t_remark = ce
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

func TestMarshal() {
	eperson := NewT_CT_Person(nil)

	e_t_name := NewT_CT_Name(nil)
	e_t_name.Base().AddXMLNS("x", "http://tutils.com")
	e_t_name.SetAttrEn("t5w0rd")

	eperson.SetElemName(e_t_name)

	e_t_remark := NewT_CT_Doc(nil)
	e_t_remark.SetText("注释")
	eperson.SetElemRemark(e_t_remark)
	eperson.SetAttrSex("male")
	eperson.SetAttrAge("18")
	eperson.SetAttrTURL("https://www.baidu.com")

	// xmlns
	eperson.Base().AddXMLNS("", "http://example.org")
	eperson.Base().AddXMLNS("e", "http://example.org")
	eperson.Base().AddXMLNS("t", "http://tutils.com")
	root := eperson.MarshalXML("http://example.org", "person")
	e_per_doc := etree.NewDocumentWithRoot(root)
	e_per_doc.Indent(4)
	e_per_doc.WriteTo(os.Stdout)

	fmt.Println("exit")
}

func TestUnmarshal() {
	e_per_doc := etree.NewDocument()
	fmt.Println(string(data.EPerson))
	fmt.Println("--------------")
	e_per_doc.ReadFromBytes(data.EPerson)

	eperson := &T_CT_Person{}
	eperson.UnmarshalXML(e_per_doc.Root())

	// root := eperson.MarshalXML("http://example.org", "person")
	// e_per_doc_2 := etree.NewDocumentWithRoot(root)
	// e_per_doc_2.Indent(4)
	// e_per_doc_2.WriteTo(os.Stdout)

	fmt.Println("exit", eperson)
}
