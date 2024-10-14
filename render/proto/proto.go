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

TODO:
1. sequence 需要重构，保证定义顺序
2. 对any和"lax"的处理

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

// AddElement 待匹配的元素只可能有两种情况：下一个元素或下一个minOccurs>0的元素
func (sl *SequenceList) AddElement(wrap *XMLElementWrap) bool {
	if len(sl.restr) == 0 || sl.MaxOccurs == 0 {
		return false
	}

	if sl.MaxOccurs > 0 && sl.occurs >= sl.MaxOccurs {
		// 超过最大重复次数
		return false
	}

	tryAdd := func(cur int) bool {
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

	if tryAdd(sl.expectIdx) {
		return true
	}

	alterExpect := sl.alterExpect()
	if alterExpect < 0 {
		return false
	}
	if tryAdd(alterExpect) {
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
