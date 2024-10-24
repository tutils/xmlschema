package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/proto"
	"github.com/tutils/xmlschema/tplcontainer"
)

type definedState int

const (
	undefined         definedState = iota // 未定义
	walkingDependency                     // 正在遍历依赖
	defined                               // 已定义
)

type TypeSymbolInfo struct {
	// Symbol   any
	FileName string
	Define   *TypeDefineBlock
}

// type TypeRecord struct {
// 	Name string
// 	NS   string

// 	ElemList *tplcontainer.SequenceMap[string, int]
// 	AttrList *tplcontainer.SequenceMap[string, int]
// }

// func NewTypeRecord(name string) *TypeRecord {
// 	return &TypeRecord{
// 		Name:     name,
// 		ElemList: tplcontainer.NewSequenceMap[string, int](),
// 		AttrList: tplcontainer.NewSequenceMap[string, int](),
// 	}
// }

// AttributeBlock 属性定义块
type AttributeBlock struct {
	Name     string
	NS       string
	TypeName string
	TypeNS   string
	Occurs   int
	Required bool
	Default  string

	CodeName     string
	CodeTypeName string
}

// ElementBlock 元素定义块
type ElementBlock struct {
	Choice    bool // 是choice
	ChoiceIdx int

	Name       string
	NS         string
	TypeName   string
	TypeNS     string
	MinOccurs  int
	MaxOccurs  int
	Occurs     int
	SimpleType bool
	Required   bool
	Path       ElementPath

	CodeName     string
	CodeTypeName string
}

// 类型定义块
type TypeDefineBlock struct {
	Name  string
	NS    string
	Attrs *tplcontainer.SequenceMap[string, *AttributeBlock]
	Elems *tplcontainer.SequenceMap[string, *ElementBlock]

	SimpleType bool
	Fallback   bool

	CodeName string
}

func NewTypeDefineBlock(name string, ns string) *TypeDefineBlock {
	return &TypeDefineBlock{
		Name:  name,
		NS:    ns,
		Attrs: tplcontainer.NewSequenceMap[string, *AttributeBlock](),
		Elems: tplcontainer.NewSequenceMap[string, *ElementBlock](),
	}
}

type Renderer struct {
	// options
	parseRecursive bool
	verbose        bool

	// runtime
	level       int
	gs          *GlobalScope
	defineState map[any]definedState

	order            *tplcontainer.SequenceMap[any, *TypeSymbolInfo]
	forwardDelcOrder *tplcontainer.SequenceMap[any, *TypeSymbolInfo]

	defBlocks      []*TypeDefineBlock
	curDef         *TypeDefineBlock
	mergeOccursSeq *proto.Sequence // 将下一个元素与父级sequence的MinOccurs和MaxOccurs合并
	// elemRequired   bool

	pathBuilder *ElementPathBuilder
}

func NewRenderer(gs *GlobalScope) *Renderer {
	return &Renderer{
		gs:               gs,
		defineState:      make(map[any]definedState),
		order:            tplcontainer.NewSequenceMap[any, *TypeSymbolInfo](),
		forwardDelcOrder: tplcontainer.NewSequenceMap[any, *TypeSymbolInfo](),

		pathBuilder: NewElementPathBuilder(),
	}
}

func (r *Renderer) prefix() string {
	var s string
	for i := 0; i < r.level; i++ {
		s += "    "
	}
	return s
}

func (r *Renderer) output(format string, a ...any) {
	if !r.verbose {
		return
	}
	format = strings.ReplaceAll(format, "```", "")
	fmt.Print(r.prefix())
	fmt.Printf(format, a...)
	fmt.Println()
}

func (r *Renderer) nextLevel() func() {
	save := r.level
	r.level++
	return func() {
		r.level = save
	}
}

func (r *Renderer) define(symbol any) func() {
	switch r.defineState[symbol] {
	case walkingDependency:
		// cycle ref
		return nil
	case defined:
		return nil
	}
	r.defineState[symbol] = walkingDependency
	return func() {
		r.defineState[symbol] = defined
	}
}

func ParseDocElement(r *Renderer, elem *etree.Element) {
	r.output("parse element: %s:%s", elem.Space, elem.Tag)
	symbol, ok := r.gs.GetElement(elem.NamespaceURI(), elem.Tag)
	if !ok {
		panic("invalid element")
	}

	defer r.nextLevel()()
	r.ParseSchemaElement(symbol.Symbol, symbol.FileName)
}

func (r *Renderer) namespace(fileName string) string {
	fs := r.gs.fileMap.MustGet(fileName)
	return fs.namespace
}

func parseAnnotaion(anno *proto.Annotation) string {
	if anno == nil {
		return ""
	}
	var sb strings.Builder
	for i, doc := range anno.DocumentationList {
		if i > 0 {
			sb.WriteString("; ")
		}
		if len(doc.Source) > 0 {
			sb.WriteString(doc.Source)
			continue
		}
		if len(doc.DivList) > 0 || len(doc.PList) > 0 {
			sb.WriteString("<HTML TEXT>")
			continue
		}
		sb.Write(doc.XMLContent)
	}
	return sb.String()
}

// (annotation?,((simpleType|complexType)?,(unique|key|keyref)*))
func (r *Renderer) ParseSchemaElement(elem *proto.Element, fileName string) {
	block := &ElementBlock{
		Occurs: 1,
		Path:   r.pathBuilder.Path(),
	}

	var prefix string
	var refElem *proto.Element
	var name string
	if len(elem.Ref) > 0 {
		p, symbol, ok := r.gs.GetElementByRef(elem, fileName)
		if !ok {
			panic("invalid element")
		}
		refElem, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
		ns := r.namespace(symbol.FileName)
		name = fmt.Sprintf("%s | %s", symbol.Symbol.Name, ns)
		block.NS = ns
	} else {
		ns := r.namespace(fileName)
		name = fmt.Sprintf("%s | %s", elem.Name, ns)
		block.NS = ns
		refElem = elem
	}
	block.Name = refElem.Name
	if refElem.MinOccurs == "" {
		block.MinOccurs = 1
	} else {
		minOccurs, err := strconv.Atoi(refElem.MinOccurs)
		if err != nil {
			panic("invalid minOccurs")
		}
		block.MinOccurs = minOccurs
	}
	if refElem.MaxOccurs == "" {
		block.MaxOccurs = 1
	} else if refElem.MaxOccurs == "unbounded" {
		block.MaxOccurs = -1
	} else {
		maxOccurs, err := strconv.Atoi(refElem.MaxOccurs)
		if err != nil {
			panic("invalid maxOccurs")
		}
		block.MaxOccurs = maxOccurs
	}
	if r.mergeOccursSeq != nil {
		if r.mergeOccursSeq.MaxOccurs == "unbounded" {
			block.MaxOccurs = -1
		}
		if r.mergeOccursSeq.MinOccurs == "0" {
			block.MinOccurs = 0
		}
		r.mergeOccursSeq = nil
	}
	lv := r.level

	elemName := refElem.Name
	for len(refElem.SubstitutionGroup) > 0 {
		_, symbol, ok := r.gs.GetElementBySubstitutionGroupRef(refElem, fileName)
		if !ok {
			panic("invalid element")
		}
		refElem, fileName = symbol.Symbol, symbol.FileName
	}

	if len(refElem.Type) > 0 {
		typePrefix, symbolType, ok := r.gs.GetTypeInFile(refElem.Type, fileName)
		if !ok {
			panic("type not found")
		}
		if len(typePrefix) > 0 {
			typePrefix += ":"
		}

		switch typ := symbolType.Symbol.(type) {
		case *proto.ComplexType:
			r.output("element: %s%s(%s%s) (%s, %s) // %s", prefix, elemName, typePrefix, typ.Name, elem.MinOccurs, elem.MaxOccurs, parseAnnotaion(refElem.Annotation))
			if r.parseRecursive {
				defer r.nextLevel()()
				r.ParseSchemaComplexType(typ, symbolType.FileName)
			}
			block.TypeName = typ.Name
			block.TypeNS = r.namespace(symbolType.FileName)

		case *proto.SimpleType:
			r.output("element: %s%s(%s%s) (%s, %s) // %s", prefix, elemName, typePrefix, typ.Name, elem.MinOccurs, elem.MaxOccurs, parseAnnotaion(refElem.Annotation))
			if r.parseRecursive {
				defer r.nextLevel()()
				r.ParseSchemaSimpleType(typ, symbolType.FileName)
			}
			block.TypeName = typ.Name
			block.TypeNS = r.namespace(symbolType.FileName)
			block.SimpleType = true
		}
	} else if refElem.ComplexType != nil {
		// 内嵌匿名类型定义
		r.output("element: %s(<complexType>) (%s, %s) // %s", refElem.Name, elem.MinOccurs, elem.MaxOccurs, parseAnnotaion(refElem.Annotation))
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaComplexType(refElem.ComplexType, fileName)
	}

	if r.curDef != nil {
		if def, ok := r.curDef.Elems.Get(name); ok {
			def.Occurs++
			r.curDef.Fallback = true
		} else {
			r.curDef.Elems.Set(name, block)
		}
	}

	if refElem.Unique != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaUnique(refElem.Unique, fileName)
	}
}

// (annotation?,(simpleContent|complexContent|((group|all|choice|sequence)?,((attribute|attributeGroup)*,anyAttribute?))))
func (r *Renderer) ParseSchemaComplexType(ct *proto.ComplexType, fileName string) {
	ns := r.namespace(fileName)
	rec := NewTypeDefineBlock(ct.Name, ns)
	r.curDef = rec

	defer func() {
		if ct.Name != "" {
			r.defBlocks = append(r.defBlocks, rec)
			r.curDef = nil
		}
	}()

	switch r.defineState[ct] {
	case walkingDependency:
		// cycle ref
		if _, ok := r.forwardDelcOrder.Get(ct.Name); !ok {
			r.forwardDelcOrder.Set(ct, &TypeSymbolInfo{
				FileName: fileName,
				Define:   &TypeDefineBlock{},
			})
		}
		return
	case defined:
		return
	}
	r.defineState[ct] = walkingDependency
	defer func() {
		r.order.Set(ct, &TypeSymbolInfo{
			FileName: fileName,
			Define:   rec,
		})
		r.defineState[ct] = defined
	}()

	r.output("complexType: %s // %s", ct.Name, parseAnnotaion(ct.Annotation))
	lv := r.level

	if ct.SimpleContent != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSimpleContent(ct.SimpleContent, fileName)
	}

	if ct.ComplexContent != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaComplexContent(ct.ComplexContent, fileName)
	}

	if ct.Group != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaGroup(ct.Group, fileName)
	}

	if ct.All != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaAll(ct.All, fileName)
	}

	if ct.Choice != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaChoice(ct.Choice, fileName)
	}

	if ct.Sequence != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSequence(ct.Sequence, fileName)
	}

	if len(ct.AttributeList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, attr := range ct.AttributeList {
			r.ParseSchemaAttribute(attr, fileName)
		}
	}

	if len(ct.AttributeGroupList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, attrGrp := range ct.AttributeGroupList {
			r.ParseSchemaAttributeGroup(attrGrp, fileName)
		}
	}
}

// (annotation?,(restriction|list|union))
func (r *Renderer) ParseSchemaSimpleType(st *proto.SimpleType, fileName string) {
	ns := r.namespace(fileName)
	rec := NewTypeDefineBlock(st.Name, ns)
	rec.SimpleType = true
	r.curDef = rec

	defer func() {
		if st.Name != "" {
			r.defBlocks = append(r.defBlocks, rec)
			r.curDef = nil
		}
	}()

	switch r.defineState[st] {
	case walkingDependency:
		// cycle ref
		if _, ok := r.forwardDelcOrder.Get(st.Name); !ok {
			r.forwardDelcOrder.Set(st, &TypeSymbolInfo{
				FileName: fileName,
				Define:   &TypeDefineBlock{},
			})
		}
		return
	case defined:
		return
	}
	r.defineState[st] = walkingDependency
	defer func() {
		r.order.Set(st, &TypeSymbolInfo{
			FileName: fileName,
			Define:   rec,
		})
		r.defineState[st] = defined
	}()

	r.output("simpleType: %s // %s", st.Name, parseAnnotaion(st.Annotation))
	lv := r.level

	if st.Restriction != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaRestriction(st.Restriction, fileName)
	}

	if st.List != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaList(st.List, fileName)
	}

	if st.Union != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaUnion(st.Union, fileName)
	}
}

// (annotation?,(simpleType?))
func (r *Renderer) ParseSchemaAttribute(attr *proto.Attribute, fileName string) {
	var prefix string
	var refAttr *proto.Attribute
	var name string
	block := &AttributeBlock{
		Occurs: 1,
	}
	if len(attr.Ref) > 0 {
		p, symbol, ok := r.gs.GetAttributeByRef(attr, fileName)
		if !ok {
			panic("invalid attribute")
		}
		refAttr, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
		ns := r.namespace(symbol.FileName)
		name = fmt.Sprintf("@%s | %s", symbol.Symbol.Name, ns)
		block.NS = ns
		if attr.Use == "required" {
			block.Required = true
		}
		if attr.Default != "" {
			block.Default = attr.Default
		} else if refAttr.Default != "" {
			block.Default = refAttr.Default
		}
	} else {
		refAttr = attr
		name = fmt.Sprintf("@%s", attr.Name)
	}
	block.Name = refAttr.Name

	if len(refAttr.Type) > 0 {
		// 存在属性类型
		typePrefix, symbolType, ok := r.gs.GetSimpleTypeInFile(refAttr.Type, fileName)
		if !ok {
			panic("simpleType not found")
		}
		if len(typePrefix) > 0 {
			typePrefix += ":"
		}

		r.output("attribute: %s%s(%s%s) // %s", prefix, refAttr.Name, typePrefix, symbolType.Symbol.Name, parseAnnotaion(refAttr.Annotation))
		if r.parseRecursive {
			defer r.nextLevel()()
			r.ParseSchemaSimpleType(symbolType.Symbol, symbolType.FileName)
		}

		block.TypeName = symbolType.Symbol.Name
		block.TypeNS = r.namespace(symbolType.FileName)
	} else if attr.SimpleType != nil {
		// 内嵌匿名类型定义
		r.output("attribute: %s(<simpleType>) // %s", attr.Name, parseAnnotaion(attr.Annotation))
		defer r.nextLevel()()
		r.ParseSchemaSimpleType(attr.SimpleType, fileName)
	}

	if r.curDef != nil {
		if def, ok := r.curDef.Attrs.Get(name); ok {
			def.Occurs++
		} else {
			r.curDef.Attrs.Set(name, block)
		}
	}
}

// (annotation?),((attribute|attributeGroup)*,anyAttribute?))
func (r *Renderer) ParseSchemaAttributeGroup(attrGrp *proto.AttributeGroup, fileName string) {
	var prefix string
	if len(attrGrp.Ref) > 0 {
		p, symbol, ok := r.gs.GetAttributeGroupByRef(attrGrp, fileName)
		if !ok {
			panic("invalid attributeGroup")
		}
		attrGrp, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}
	r.output("attributeGroup: %s%s // %s", prefix, attrGrp.Name, parseAnnotaion(attrGrp.Annotation))

	defer r.nextLevel()()
	for _, attr := range attrGrp.AttributeList {
		r.ParseSchemaAttribute(attr, fileName)
	}
	for _, attrGrp := range attrGrp.AttributeGroupList {
		r.ParseSchemaAttributeGroup(attrGrp, fileName)
	}
}

// // (annotation?,(element|group|choice|sequence|any)*)
// func (r *Renderer) ParseSchemaSequence_old(seq *proto.Sequence, fileName string) {
// 	r.output("sequence (%s, %s) // %s", seq.MinOccurs, seq.MaxOccurs, parseAnnotaion(seq.Annotation))
// 	lv := r.level

// 	if len(seq.ElementList) > 0 {
// 		r.level = lv
// 		defer r.nextLevel()()
// 		for _, elem := range seq.ElementList {
// 			r.ParseSchemaElement(elem, fileName)
// 		}
// 	}

// 	if len(seq.GroupList) > 0 {
// 		r.level = lv
// 		defer r.nextLevel()()
// 		for _, grp := range seq.GroupList {
// 			r.ParseSchemaGroup(grp, fileName)
// 		}
// 	}

// 	if len(seq.ChoiceList) > 0 {
// 		r.level = lv
// 		defer r.nextLevel()()
// 		for _, ch := range seq.ChoiceList {
// 			r.ParseSchemaChoice(ch, fileName)
// 		}
// 	}

// 	if seq.Sequence != nil {
// 		r.level = lv
// 		defer r.nextLevel()()
// 		r.ParseSchemaSequence(seq.Sequence, fileName)
// 	}

// 	if len(seq.AnyList) > 0 {
// 		r.level = lv
// 		defer r.nextLevel()()
// 		for _, an := range seq.AnyList {
// 			r.ParseSchemaAny(an, fileName)
// 		}
// 	}
// }

// var chrec = make(map[*proto.Choice]int)
// var chrec2 = make(map[*TypeDefineBlock]int)

// (annotation?,(element|group|choice|sequence|any)*)
func (r *Renderer) ParseSchemaSequence(seq *proto.Sequence, fileName string) {
	r.pathBuilder.Push(seq)
	defer r.pathBuilder.Pop()

	r.output("sequence (%s, %s) // %s", seq.MinOccurs, seq.MaxOccurs, parseAnnotaion(seq.Annotation))
	lv := r.level

	if seq.MaxOccurs != "" && seq.MaxOccurs != "1" {
		// 多次出现的sequence
		r.curDef.Fallback = true
		if seq.MaxOccurs == "unbounded" && len(seq.NestedParticleList) == 1 {
			if _, ok := seq.NestedParticleList[0].(*proto.Element); ok {
				r.mergeOccursSeq = seq
				r.curDef.Fallback = false
			}
		}
	}

	for _, part := range seq.NestedParticleList {
		switch ch := part.(type) {
		case *proto.Element:
			r.level = lv
			func() {
				defer r.nextLevel()()
				r.ParseSchemaElement(ch, fileName)
			}()
		case *proto.Group:
			r.level = lv
			func() {
				defer r.nextLevel()()
				r.ParseSchemaGroup(ch, fileName)
			}()
		case *proto.Choice:
			r.level = lv
			func() {
				defer r.nextLevel()()
				r.ParseSchemaChoice(ch, fileName)
			}()
		case *proto.Sequence:
			r.level = lv
			func() {
				defer r.nextLevel()()
				r.ParseSchemaSequence(ch, fileName)
			}()
		case *proto.Any:
			r.level = lv
			func() {
				defer r.nextLevel()()
				r.ParseSchemaAny(ch, fileName)
			}()
		}
	}
}

// (annotation?,(all|choice|sequence)?)
func (r *Renderer) ParseSchemaGroup(grp *proto.Group, fileName string) {
	var prefix string
	var refGrp *proto.Group
	if len(grp.Ref) > 0 {
		p, symbol, ok := r.gs.GetGroupByRef(grp, fileName)
		if !ok {
			panic("invalid attributeGroup")
		}
		refGrp, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	} else {
		refGrp = grp
	}
	r.output("group: %s%s (%s, %s) // %s", prefix, refGrp.Name, grp.MinOccurs, grp.MaxOccurs, parseAnnotaion(refGrp.Annotation))
	lv := r.level

	if refGrp.Choice != nil {
		defer r.nextLevel()()
		r.ParseSchemaChoice(refGrp.Choice, fileName)
	}

	if refGrp.Sequence != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSequence(refGrp.Sequence, fileName)
	}
}

// (annotation?,(element|group|choice|sequence|any)*)
func (r *Renderer) ParseSchemaChoice(ch *proto.Choice, fileName string) {
	r.pathBuilder.Push(ch)
	defer r.pathBuilder.Pop()

	r.output("choice (%s, %s) // %s", ch.MinOccurs, ch.MaxOccurs, parseAnnotaion(ch.Annotation))
	lv := r.level

	if len(ch.ElementList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, elem := range ch.ElementList {
			r.ParseSchemaElement(elem, fileName)
		}
	}

	if len(ch.GroupList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, grp := range ch.GroupList {
			r.ParseSchemaGroup(grp, fileName)
		}
	}
	if ch.Choice != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaChoice(ch.Choice, fileName)
	}

	if ch.Sequence != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSequence(ch.Sequence, fileName)
	}

	if ch.Any != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaAny(ch.Any, fileName)
	}
}

// (annotation?,element*)
func (r *Renderer) ParseSchemaAll(all *proto.All, fileName string) {
	r.output("all (%s,)", all.MinOccurs)

	if len(all.ElementList) > 0 {
		defer r.nextLevel()()
		for _, elem := range all.ElementList {
			r.ParseSchemaElement(elem, fileName)
		}
	}
}

// (annotation?,(restriction|extension))
func (r *Renderer) ParseSchemaSimpleContent(sc *proto.SimpleContent, fileName string) {
	r.output("simpleContent")
	lv := r.level

	if sc.Restriction != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaRestriction(sc.Restriction, fileName)
	}

	if sc.Extension != nil {
		defer r.nextLevel()()
		r.ParseSchemaExtension(sc.Extension, fileName)
	}
}

// (annotation?,(restriction|extension))
func (r *Renderer) ParseSchemaComplexContent(cc *proto.ComplexContent, fileName string) {
	r.output("complexContent")
	lv := r.level

	if cc.Restriction != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaRestriction(cc.Restriction, fileName)
	}

	if cc.Extension != nil {
		defer r.nextLevel()()
		r.ParseSchemaExtension(cc.Extension, fileName)
	}
}

// (annotation?)
func (r *Renderer) ParseSchemaAny(an *proto.Any, fileName string) {
	r.output("any (%s, %s)", an.MinOccurs, an.MaxOccurs)
}

// (annotation?,((group|all|choice|sequence)?,((attribute|attributeGroup)*,anyAttribute?)))
func (r *Renderer) ParseSchemaExtension(ext *proto.Extension, fileName string) {
	if len(ext.Base) == 0 {
		panic("invalid extension")
	}
	r.output("extension: %s", ext.Base)
	lv := r.level

	_, symbolType, ok := r.gs.GetTypeInFile(ext.Base, fileName)
	if !ok {
		panic("type not found")
	}
	switch typ := symbolType.Symbol.(type) {
	case *proto.ComplexType:
		if r.parseRecursive {
			r.ParseSchemaComplexType(typ, symbolType.FileName)
		}

	case *proto.SimpleType:
		if r.parseRecursive {
			r.ParseSchemaSimpleType(typ, symbolType.FileName)
		}
	}

	if ext.Sequence != nil {
		defer r.nextLevel()()
		r.ParseSchemaSequence(ext.Sequence, fileName)
	}

	if ext.Choice != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaChoice(ext.Choice, fileName)
	}

	if ext.Group != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaGroup(ext.Group, fileName)
	}

	if len(ext.AttributeList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, attr := range ext.AttributeList {
			r.ParseSchemaAttribute(attr, fileName)
		}
	}

	if len(ext.AttributeGroupList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, attrGrp := range ext.AttributeGroupList {
			r.ParseSchemaAttributeGroup(attrGrp, fileName)
		}
	}
}

// simpleType: (annotation?,(simpleType?,(minExclusive|minInclusive|maxExclusive|maxInclusive|totalDigits|fractionDigits|length|minLength|maxLength|enumeration|whiteSpace|pattern)*))
// simpleContent: (annotation?,(simpleType?,(minExclusive |minInclusive|maxExclusive|maxInclusive|totalDigits|fractionDigits|length|minLength|maxLength|enumeration|whiteSpace|pattern)*)?,((attribute|attributeGroup)*,anyAttribute?))
// complexContent: (annotation?,(group|all|choice|sequence)?,((attribute|attributeGroup)*,anyAttribute?))
func (r *Renderer) ParseSchemaRestriction(restr *proto.Restriction, fileName string) {
	lv := r.level
	if len(restr.Base) > 0 {
		r.output("restriction: %s // %s", restr.Base, parseAnnotaion(restr.Annotation))
	} else if restr.SimpleType != nil {
		r.output("restriction: <simpleType> // %s", restr.Base, parseAnnotaion(restr.Annotation))
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSimpleType(restr.SimpleType, fileName)
	} else {
		panic("invalid restriction")
	}

	if len(restr.EnumerationList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, enum := range restr.EnumerationList {
			r.ParseSchemaEnumeration(enum, fileName)
		}
	}

	if restr.Group != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaGroup(restr.Group, fileName)
	}

	if restr.Sequence != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSequence(restr.Sequence, fileName)
	}

	if len(restr.AttributeList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, attr := range restr.AttributeList {
			r.ParseSchemaAttribute(attr, fileName)
		}
	}
}

func (r *Renderer) ParseSchemaEnumeration(enum *proto.Enumeration, fileName string) {
	r.output("enumeration: %s // %s", enum.Value, parseAnnotaion(enum.Annotation))
}

// (annotation?,(simpleType*))
func (r *Renderer) ParseSchemaUnion(uni *proto.Union, fileName string) {
	r.output("union")
	lv := r.level

	if len(uni.MemberTypes) > 0 {
		sli := strings.Split(uni.MemberTypes, " ")
		if len(sli) == 0 {
			panic("empty memberTypes")
		}

		defer r.nextLevel()()
		for _, typ := range sli {
			_, symbolType, ok := r.gs.GetTypeInFile(typ, fileName)
			if !ok {
				panic("type not found")
			}
			switch typ := symbolType.Symbol.(type) {
			case *proto.ComplexType:
				if r.parseRecursive {
					r.ParseSchemaComplexType(typ, symbolType.FileName)
				}

			case *proto.SimpleType:
				if r.parseRecursive {
					r.ParseSchemaSimpleType(typ, symbolType.FileName)
				}
			}
		}
	}

	if len(uni.SimpleTypeList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, st := range uni.SimpleTypeList {
			r.ParseSchemaSimpleType(st, fileName)
		}
	}
}

// (annotation?,(simpleType?))
func (r *Renderer) ParseSchemaList(lst *proto.List, fileName string) {
	r.output("list")
	lv := r.level

	if len(lst.ItemType) > 0 {
		_, symbolType, ok := r.gs.GetTypeInFile(lst.ItemType, fileName)
		if !ok {
			panic("type not found")
		}
		switch typ := symbolType.Symbol.(type) {
		case *proto.ComplexType:
			if r.parseRecursive {
				r.ParseSchemaComplexType(typ, symbolType.FileName)
			}

		case *proto.SimpleType:
			if r.parseRecursive {
				r.ParseSchemaSimpleType(typ, symbolType.FileName)
			}
		}
	}

	if lst.SimpleType != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSimpleType(lst.SimpleType, fileName)
	}
}

// (annotation?,(selector,field+))
func (r *Renderer) ParseSchemaUnique(uniq *proto.Unique, fileName string) {
	r.output("unique")

	if uniq.Selector != nil {
		r.output("selector: %s", uniq.Selector.Xpath)
	}

	if uniq.Field != nil {
		r.output("field: %s", uniq.Field.Xpath)
	}
}
