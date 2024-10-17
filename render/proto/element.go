package proto

import (
	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/tplcontainer"
)

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
