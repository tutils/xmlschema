package render

import (
	"fmt"
	"strings"

	"github.com/tutils/xmlschema/proto"
	"github.com/tutils/xmlschema/tplcontainer"
)

const (
	isMutexFlag uint8 = 1 << 7
	isArrayFlag uint8 = 1 << 6
	indexMask   uint8 = 0b00111111
)

type NodeProperty uint8

func NewNodeProperty(index int, isMutex bool, isArray bool) NodeProperty {
	if index >= (1 << 6) {
		panic("invalid index")
	}
	var n uint8 = uint8(index)
	if isMutex {
		n |= isMutexFlag
	}
	if isArray {
		n |= isArrayFlag
	}
	return NodeProperty(n)
}

func (n NodeProperty) IsMutex() bool {
	return (uint8(n) & isMutexFlag) != 0
}

func (n NodeProperty) IsArray() bool {
	return (uint8(n) & isArrayFlag) != 0
}

func (n NodeProperty) Index() int {
	return int(uint8(n) & indexMask)
}

func (n NodeProperty) UInt8() uint8 {
	return uint8(n)
}

type ElementPath []NodeProperty

func (p *ElementPath) AddNode(index int, isMutex bool, isArray bool) {
	*p = append(*p, NewNodeProperty(index, isMutex, isArray))
}
func (p *ElementPath) AddNodeRaw(n NodeProperty) {
	*p = append(*p, n)
}

func (p ElementPath) MutexWith(p2 ElementPath) bool {
	n := len(p)
	n2 := len(p2)
	i := 0
	for ; i < n && i < n2 && p[i] == p2[i]; i++ {
	}

	if i == 0 {
		return false
	}

	return NodeProperty(p[i-1]).IsMutex()
}

// func (p ElementPath) IsArray() bool {
// 	sz := len(p)
// 	if sz == 0 {
// 		return false
// 	}
// 	return p[sz-1].IsArray()
// }

func (p ElementPath) String() string {
	var sb strings.Builder
	for _, n := range p {
		sb.WriteByte(byte('/'))
		if n.IsArray() {
			sb.WriteString("[]")
		}
		if n.IsMutex() {
			sb.WriteString(fmt.Sprintf("c%d", n.Index()))
		} else {
			sb.WriteString(fmt.Sprintf("s%d", n.Index()))
		}
	}
	return sb.String()
}

type ElementPathBuilder struct {
	stack         *tplcontainer.Stack[NodeProperty]
	sequenceIndex int
	choiceIndex   int
}

func NewElementPathBuilder() *ElementPathBuilder {
	return &ElementPathBuilder{
		stack: tplcontainer.NewStack[NodeProperty](),
	}
}

func (b *ElementPathBuilder) Push(e any) {
	switch typ := e.(type) {
	case *proto.Sequence:
		b.pushSequence(typ)
	case *proto.Choice:
		b.pushChoice(typ)
	default:
		panic("invalid type")
	}
}

func (b *ElementPathBuilder) pushSequence(seq *proto.Sequence) {
	maxOccurs := seq.MaxOccurs
	// var isArray bool
	// if top, ok := b.stack.Top(); ok {
	// 	isArray = top.IsArray() || ((maxOccurs != "") && (maxOccurs != "1"))
	// } else {
	// 	isArray = (maxOccurs != "") && (maxOccurs != "1")
	// }
	isArray := (maxOccurs != "") && (maxOccurs != "1")

	n := NewNodeProperty(b.sequenceIndex, false, isArray)
	b.stack.Push(n)
	b.sequenceIndex++
}

func (b *ElementPathBuilder) pushChoice(ch *proto.Choice) {
	maxOccurs := ch.MaxOccurs
	// var isArray bool
	// if top, ok := b.stack.Top(); ok {
	// 	isArray = top.IsArray() || ((maxOccurs != "") && (maxOccurs != "1"))
	// } else {
	// 	isArray = (maxOccurs != "") && (maxOccurs != "1")
	// }
	isArray := (maxOccurs != "") && (maxOccurs != "1")

	n := NewNodeProperty(b.choiceIndex, true, isArray)
	b.stack.Push(n)
	b.choiceIndex++
}

func (b *ElementPathBuilder) Pop() {
	b.stack.Pop()
	if b.stack.Len() == 0 {
		b.sequenceIndex = 0
		b.choiceIndex = 0
	}
}

func (b *ElementPathBuilder) Path() ElementPath {
	var path ElementPath
	for it := b.stack.BottomIterator(); it.Valid(); it = it.Next() {
		path.AddNodeRaw(it.Value())
	}
	return path
}
