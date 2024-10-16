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

group定义、group/sequence、group/choice 不允许出现minOccurs或maxOccurs属性
group引用允许出现minOccurs或maxOccurs属性

对于sequence引用的多个group中存在同名的元素的情况，对于可能造成歧义的可选元素，一律按照本轮排序最前的group算，不会回溯判断

如果sequence.maxOccurs > 1 全部使用Add方法，否则使用Set方法

group直接展开后，是否有重名元素
	1. 无：每个元素渲染set/get方法
	2. 有：使用兜底结构（xml基本操作方法）

sequence
	1. maxOccurs<=1：每个元素渲染set/get方法
	2. maxOccurs>1
		2.1 子元素只有一个：等价于一个元素的数组
		2.2 子元素不止一个：使用兜底结构（xml基本操作方法）

TODO:
1. 对any和"lax"的处理

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

// var _ XMLElement = (*E_CT_Work)(nil)

// type E_CT_Work struct {
// 	base *XMLElementBase

// 	_seqLst []any

// 	// AG_WhereAndDuration
// 	a_name     string
// 	a_duration uint16
// }

// // Base implements XMLElement.
// func (e *E_CT_Work) Base() *XMLElementBase {
// 	panic("unimplemented")
// }

// // MarshalXML implements XMLElement.
// func (e *E_CT_Work) MarshalXML(ns string, tag string) *etree.Element {
// 	panic("unimplemented")
// }

// // UnmarshalXML implements XMLElement.
// func (e *E_CT_Work) UnmarshalXML(ee *etree.Element) {
// 	panic("unimplemented")
// }

// func (e *E_CT_Work) SetAttrRemark(v string) {
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

type XMLElementWrap struct {
	NS  string
	Tag string
	E   XMLElement
}

func NewElementWrap(ns string, tag string, e XMLElement) *XMLElementWrap {
	return &XMLElementWrap{
		NS:  ns,
		Tag: tag,
		E:   e,
	}
}

type XMLElementRestriction struct {
	NS        string
	Tag       string
	MinOccurs int
	MaxOccurs int // unbounded: -1
}

func NewXMLElementRestriction(ns string, tag string, min int, max int) *XMLElementRestriction {
	return &XMLElementRestriction{
		NS:        ns,
		Tag:       tag,
		MinOccurs: min,
		MaxOccurs: max,
	}
}

type SequenceList struct {
	base *XMLElementBase

	restr     []*XMLElementRestriction
	expectIdx int // expect
	occurs    int // 遍历过的轮数
	MinOccurs int
	MaxOccurs int // unbounded: -1

	list []*XMLElementWrap
}

func NewSequenceList(base *XMLElementBase, restr []*XMLElementRestriction, min int, max int) *SequenceList {
	return &SequenceList{
		base:      base,
		restr:     restr,
		MinOccurs: min,
		MaxOccurs: max,
	}
}

func (sl *SequenceList) Base() *XMLElementBase {
	return sl.base
}

// alterExpect 返回下备选expect元素位置，即下一个minOccurs>0的元素，如果没有下一个元素位置返回-1
// 可以通过比较返回的元素位置和当前位置来确定是否进入了下一轮
func (sl *SequenceList) alterExpect() int {
	first := true
	for cur := sl.expectIdx; ; {
		if cur == sl.expectIdx {
			if !first {
				// 转回来了
				return -1
			}
			first = false
		} else {
			if sl.restr[cur].MinOccurs == 0 {
				return cur
			}
		}

		if cur >= len(sl.restr)-1 {
			// 准备进入下一轮
			if sl.MaxOccurs >= 0 && sl.occurs >= sl.MaxOccurs {
				return -1
			}
			cur = 0
		} else {
			cur++
		}
	}
}

func (sl *SequenceList) tryAdd(cur int, wrap *XMLElementWrap) bool {
	// 已经确定可以添加，这里主要进行sequence整体次数判断
	if sl.restr[cur].NS != wrap.NS || sl.restr[cur].Tag != wrap.Tag {
		// 匹配当前expect
		return false
	}
	if cur >= len(sl.restr)-1 {
		sl.occurs++
		sl.expectIdx = 0
	} else {
		sl.expectIdx = cur + 1
	}
	wrap.E.Base().SetParent(sl.Base())
	sl.list = append(sl.list, wrap)
	return true
}

// AddElement 待匹配的元素只可能有两种情况：下一个元素或下一个minOccurs>0的元素
func (sl *SequenceList) AddElement(wrap *XMLElementWrap) bool {
	if len(sl.restr) == 0 || sl.MaxOccurs == 0 {
		return false
	}

	if sl.MaxOccurs > 0 && sl.occurs >= sl.MaxOccurs {
		// 超过最大重复次数
		return false
	}

	if sl.tryAdd(sl.expectIdx, wrap) {
		return true
	}

	alterExpect := sl.alterExpect()
	if alterExpect < 0 {
		return false
	}
	if sl.tryAdd(alterExpect, wrap) {
		return true
	}
	return false
}

func (sl *SequenceList) UnmarshalXML(base *XMLElementBase, ee *etree.Element) {
	// for {

	// }

	// ce := NewT_CT_Name(base)
	// ce.UnmarshalXML(cee)
	// e.seqLst = append(e.seqLst, NewElementWrap(
	// 	"http://tutils.com",
	// 	"adoc",
	// 	ce,
	// ))

	// if base.VarifyETreeTag(
	// 	cee,
	// 	"http://tutils.com",
	// 	"adoc",
	// ) {

	// 	continue
	// }
}

func (sl *SequenceList) MarshalXML(ee *etree.Element) {
	for _, ceInfo := range sl.list {
		cee := ceInfo.E.MarshalXML(ceInfo.NS, ceInfo.Tag)
		ee.AddChild(cee)
	}
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

	var e_t_arrName *T_CT_Name
	e_t_arrName = NewT_CT_Name(nil)
	e_t_arrName.SetAttrEn("xxx")
	eperson.AddElemArrName(e_t_arrName)
	e_t_arrName = NewT_CT_Name(nil)
	e_t_arrName.SetAttrEn("yyy")
	eperson.AddElemArrName(e_t_arrName)

	e_t_defbyseq := NewT_CT_DefBySeq(nil)
	var doc *T_CT_Doc
	doc = NewT_CT_Doc(nil)
	doc.SetText("aaa")
	e_t_defbyseq.AddElemA(doc)
	doc = NewT_CT_Doc(nil)
	doc.SetText("ccc")
	e_t_defbyseq.AddElemC(doc)
	doc = NewT_CT_Doc(nil)
	doc.SetText("bbb")
	e_t_defbyseq.AddElemB(doc)
	doc = NewT_CT_Doc(nil)
	doc.SetText("ddd")
	e_t_defbyseq.AddElemD(doc)
	eperson.SetElemDefbyseq(e_t_defbyseq)

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

	eperson := NewT_CT_Person(nil)
	eperson.UnmarshalXML(e_per_doc.Root())

	root := eperson.MarshalXML("http://example.org", "person")
	e_per_doc_2 := etree.NewDocumentWithRoot(root)
	e_per_doc_2.Indent(4)
	e_per_doc_2.WriteTo(os.Stdout)

	fmt.Println("exit", eperson)
}
