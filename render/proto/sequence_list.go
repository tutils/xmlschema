package proto

import "github.com/beevik/etree"

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
