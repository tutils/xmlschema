package render

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/boot"
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
	Define   *boot.TypeDefineBlock
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

	parsingType bool
}

func NewRenderer(gs *GlobalScope) *Renderer {
	return &Renderer{
		gs:               gs,
		defineState:      make(map[any]definedState),
		order:            tplcontainer.NewSequenceMap[any, *TypeSymbolInfo](),
		forwardDelcOrder: tplcontainer.NewSequenceMap[any, *TypeSymbolInfo](),
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

func (r *Renderer) parseAnnotaion(anno *proto.Annotation) string {
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
	var prefix string
	if len(elem.Ref) > 0 {
		p, symbol, ok := r.gs.GetElementByRef(elem, fileName)
		if !ok {
			panic("invalid element")
		}
		elem, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}
	lv := r.level

	if len(elem.Type) > 0 {
		typePrefix, symbolType, ok := r.gs.GetTypeInFile(elem.Type, fileName)
		if !ok {
			panic("type not found")
		}
		if len(typePrefix) > 0 {
			typePrefix += ":"
		}

		switch typ := symbolType.Symbol.(type) {
		case *proto.ComplexType:
			r.output("element: %s%s(%s%s) // %s", prefix, elem.Name, typePrefix, typ.Name, r.parseAnnotaion(elem.Annotation))

			defer r.nextLevel()()
			r.ParseSchemaComplexType(typ, symbolType.FileName)

		case *proto.SimpleType:
			r.output("element: %s%s(%s%s) // %s", prefix, elem.Name, typePrefix, typ.Name, r.parseAnnotaion(elem.Annotation))

			defer r.nextLevel()()
			r.ParseSchemaSimpleType(typ, symbolType.FileName)
		}
	} else if elem.ComplexType != nil {
		r.output("element: %s(<complexType>) // %s", elem.Name, r.parseAnnotaion(elem.Annotation))
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaComplexType(elem.ComplexType, fileName)
	}

	if elem.Unique != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaUnique(elem.Unique, fileName)
	}
}

// (annotation?,(simpleContent|complexContent|((group|all|choice|sequence)?,((attribute|attributeGroup)*,anyAttribute?))))
func (r *Renderer) ParseSchemaComplexType(ct *proto.ComplexType, fileName string) {
	if !r.parseRecursive {
		if (r.parsingType || r.level > 0) && len(ct.Name) > 0 {
			return
		}
		r.parsingType = true
		defer func() {
			r.parsingType = false
		}()
	}

	switch r.defineState[ct] {
	case walkingDependency:
		// cycle ref
		if _, ok := r.forwardDelcOrder.Get(ct.Name); !ok {
			r.forwardDelcOrder.Set(ct, &TypeSymbolInfo{
				FileName: fileName,
				Define:   &boot.TypeDefineBlock{},
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
			Define:   &boot.TypeDefineBlock{},
		})
		r.defineState[ct] = defined
	}()

	r.output("complexType: %s // %s", ct.Name, r.parseAnnotaion(ct.Annotation))
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
	if !r.parseRecursive {
		if (r.parsingType || r.level > 0) && len(st.Name) > 0 {
			return
		}
		r.parsingType = true
		defer func() {
			r.parsingType = false
		}()
	}

	switch r.defineState[st] {
	case walkingDependency:
		// cycle ref
		if _, ok := r.forwardDelcOrder.Get(st.Name); !ok {
			r.forwardDelcOrder.Set(st, &TypeSymbolInfo{
				FileName: fileName,
				Define:   &boot.TypeDefineBlock{},
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
			Define:   &boot.TypeDefineBlock{},
		})
		r.defineState[st] = defined
	}()

	r.output("simpleType: %s // %s", st.Name, r.parseAnnotaion(st.Annotation))
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
	if len(attr.Ref) > 0 {
		p, symbol, ok := r.gs.GetAttributeByRef(attr, fileName)
		if !ok {
			panic("invalid attribute")
		}
		attr, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}

	if len(attr.Type) > 0 {
		typePrefix, symbolType, ok := r.gs.GetSimpleTypeInFile(attr.Type, fileName)
		if !ok {
			panic("simpleType not found")
		}

		r.output("attribute: %s%s(%s%s) // %s", prefix, attr.Name, typePrefix, symbolType.Symbol.Name, r.parseAnnotaion(attr.Annotation))
		defer r.nextLevel()()
		r.ParseSchemaSimpleType(symbolType.Symbol, symbolType.FileName)
		return
	}

	if attr.SimpleType != nil {
		r.output("attribute: %s(<simpleType>) // %s", attr.Name, r.parseAnnotaion(attr.Annotation))
		defer r.nextLevel()()
		r.ParseSchemaSimpleType(attr.SimpleType, fileName)
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
	r.output("attributeGroup: %s%s // %s", prefix, attrGrp.Name, r.parseAnnotaion(attrGrp.Annotation))

	defer r.nextLevel()()
	for _, attr := range attrGrp.AttributeList {
		r.ParseSchemaAttribute(attr, fileName)
	}
	for _, attrGrp := range attrGrp.AttributeGroupList {
		r.ParseSchemaAttributeGroup(attrGrp, fileName)
	}
}

// (annotation?,(element|group|choice|sequence|any)*)
func (r *Renderer) ParseSchemaSequence(seq *proto.Sequence, fileName string) {
	r.output("sequence // %s", r.parseAnnotaion(seq.Annotation))
	lv := r.level

	if len(seq.ElementList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, elem := range seq.ElementList {
			r.ParseSchemaElement(elem, fileName)
		}
	}

	if len(seq.GroupList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, grp := range seq.GroupList {
			r.ParseSchemaGroup(grp, fileName)
		}
	}

	if len(seq.ChoiceList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, ch := range seq.ChoiceList {
			r.ParseSchemaChoice(ch, fileName)
		}
	}

	if seq.Sequence != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSequence(seq.Sequence, fileName)
	}

	if len(seq.AnyList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, an := range seq.AnyList {
			r.ParseSchemaAny(an, fileName)
		}
	}
}

// (annotation?,(all|choice|sequence)?)
func (r *Renderer) ParseSchemaGroup(grp *proto.Group, fileName string) {
	var prefix string
	if len(grp.Ref) > 0 {
		p, symbol, ok := r.gs.GetGroupByRef(grp, fileName)
		if !ok {
			panic("invalid attributeGroup")
		}
		grp, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}
	r.output("group: %s%s // %s", prefix, grp.Name, r.parseAnnotaion(grp.Annotation))
	lv := r.level

	if grp.Choice != nil {
		defer r.nextLevel()()
		r.ParseSchemaChoice(grp.Choice, fileName)
	}

	if grp.Sequence != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSequence(grp.Sequence, fileName)
	}
}

// (annotation?,(element|group|choice|sequence|any)*)
func (r *Renderer) ParseSchemaChoice(ch *proto.Choice, fileName string) {
	r.output("choice // %s", r.parseAnnotaion(ch.Annotation))
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
	r.output("all")

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
	r.output("any")
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
		r.ParseSchemaComplexType(typ, symbolType.FileName)

	case *proto.SimpleType:

		r.ParseSchemaSimpleType(typ, symbolType.FileName)
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
func (r *Renderer) ParseSchemaRestriction(rest *proto.Restriction, fileName string) {
	lv := r.level
	if len(rest.Base) > 0 {
		r.output("restriction: %s // %s", rest.Base, r.parseAnnotaion(rest.Annotation))
	} else if rest.SimpleType != nil {
		r.output("restriction: <simpleType> // %s", rest.Base, r.parseAnnotaion(rest.Annotation))
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSimpleType(rest.SimpleType, fileName)
	} else {
		panic("invalid restriction")
	}

	if len(rest.EnumerationList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, enum := range rest.EnumerationList {
			r.ParseSchemaEnumeration(enum, fileName)
		}
	}

	if rest.Group != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaGroup(rest.Group, fileName)
	}

	if rest.Sequence != nil {
		r.level = lv
		defer r.nextLevel()()
		r.ParseSchemaSequence(rest.Sequence, fileName)
	}

	if len(rest.AttributeList) > 0 {
		r.level = lv
		defer r.nextLevel()()
		for _, attr := range rest.AttributeList {
			r.ParseSchemaAttribute(attr, fileName)
		}
	}
}

func (r *Renderer) ParseSchemaEnumeration(enum *proto.Enumeration, fileName string) {
	r.output("enumeration: %s // %s", enum.Value, r.parseAnnotaion(enum.Annotation))
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
				r.ParseSchemaComplexType(typ, symbolType.FileName)

			case *proto.SimpleType:
				r.ParseSchemaSimpleType(typ, symbolType.FileName)
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
			r.ParseSchemaComplexType(typ, symbolType.FileName)

		case *proto.SimpleType:
			r.ParseSchemaSimpleType(typ, symbolType.FileName)
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

// 遍历某个文件夹下的全部schema文件
func LoadAllSchema() {
	gs := NewGlobalScope()
	// fs := gs.LoadSchema("data/schemas/ooxml/pml-slide.xsd")
	// prefix, symb, ok := fs.GetElement("p:sld")
	// fmt.Println(prefix, symb, ok)
	// fs = gs.LoadSchema("data/schemas/ooxml/pml-presentation.xsd")
	// prefix, symb, ok = fs.GetElement("p:presentation")
	// fmt.Println(prefix, symb, ok)
	// gs.LoadSchema(XSSchemaLocation)
	// gs.LoadSchema(XMLSchemaLocation)

	gs.LoadSchemaFilesFromDirectory("data/schemas/ooxml")
	for _, name := range gs.fileMap.Order() {
		fs := gs.fileMap.MustGet(name)
		if len(fs.schema.AttributeList) != 0 {
			fmt.Println(fs.name)
		}
	}

	// symbol, ok := gs.GetElement(XSNamespace, "element")
	// if !ok {
	// 	panic("symbol not found")
	// }

	// r := newParseContext(gs)
	// ParseSchemaElement(symbol.Symbol, symbol.FileName)

	zr, err := zip.OpenReader("列表样式还原- Case 文件 .pptx")
	if err != nil {
		panic(err)
	}
	defer zr.Close()
	for _, file := range zr.Reader.File {
		fp, err := file.Open()
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		// fmt.Println(file.Name)
		if file.Name == "ppt/slides/slide2.xml" {
			TestParse(fp, gs)
		}
	}

	fmt.Println("exit.")
}

func escapedStr(s string) string {
	var escapedStr string
	for _, r := range s {
		// 检查是否是高代理项的 Unicode 字符
		if r >= 0xD800 && r <= 0xDFFF {
			// 处理 UTF-16 代理对
			r1, r2 := utf16.EncodeRune(r)
			escapedStr += fmt.Sprintf("\\u%04x\\u%04x", r1, r2)
		} else {
			escapedStr += fmt.Sprintf("\\u%04x", r)
		}
	}
	return escapedStr
}

func TestParse(r io.ReadCloser, gs *GlobalScope) {
	doc := etree.NewDocument()
	_, err := doc.ReadFrom(r)
	if err != nil {
		panic(err)
	}

	root := doc.Root()
	root.ChildElements()

	prefixMap := make(map[string]string)
	for _, attr := range root.Attr {
		if attr.Space == "xmlns" {
			prefixMap[attr.Key] = attr.Value
		}
	}

	orderCtx := NewRenderer(gs)
	orderCtx.parseRecursive = true
	orderCtx.verbose = false
	ParseDocElement(orderCtx, root)
	sli := root.FindElements("//p:sld/p:cSld/p:spTree/p:sp/p:txBody/a:p/a:pPr[a:buChar]")
	fmt.Println(sli)
	for _, e := range sli {
		buFont := e.SelectElement("a:buFont")
		buChar := e.SelectElement("a:buChar")
		fmt.Println("typeface:", buFont.SelectAttr("typeface").Value)
		ch := buChar.SelectAttr("char").Value
		fmt.Println("char:", escapedStr(ch))
	}

	genCtx := NewRenderer(gs)
	genCtx.parseRecursive = false
	genCtx.verbose = true
	for _, symb := range orderCtx.order.Order() {
		info := orderCtx.order.MustGet(symb)
		switch typ := symb.(type) {
		case *proto.ComplexType:
			genCtx.ParseSchemaComplexType(typ, info.FileName)
		case *proto.SimpleType:
			genCtx.ParseSchemaSimpleType(typ, info.FileName)
		}
	}

	// fmt.Println(doc, root)
	// var root boot.XMLElement
	// dec := xml.NewDecoder(fp)
	// err := dec.Decode(&root)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(root)

	fmt.Println("debug")
}

func GenAllSymbolText() {
	gs := NewGlobalScope()
	gs.LoadSchemaFilesFromDirectory("data/schemas/ooxml")
	for _, name := range gs.fileMap.Order() {
		fs := gs.fileMap.MustGet(name)
		if len(fs.schema.AttributeList) != 0 {
			fmt.Println(fs.name)
		}
	}

	// // 生成定义顺序
	// orderCtx := newParseContext(gs)
	// orderCtx.parseRecursive = true
	// orderCtx.verbose = false

	// // 按顺序生成
	// genCtx := newParseContext(gs)
	// genCtx.parseRecursive = false
	// genCtx.verbose = true
	// for _, symb := range orderCtx.order.Order() {
	// 	info := orderCtx.order.MustGet(symb)
	// 	switch typ := symb.(type) {
	// 	case *proto.ComplexType:
	// 		r.ParseSchemaComplexType(typ, info.FileName)
	// 	case *proto.SimpleType:
	// 		r.ParseSchemaSimpleType(typ, info.FileName)
	// 	}
	// }

	// 按命名空间生成
	genCtx := NewRenderer(gs)
	genCtx.parseRecursive = false
	genCtx.verbose = true

	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()

	for _, ns := range gs.namespaceMap.Order() {
		u, err := url.Parse(ns)
		if err != nil {
			panic(err)
		}

		baseDir := filepath.Join("output", strings.ReplaceAll(strings.ReplaceAll(u.Path, "/", "_"), ".", "_"))
		symbs := gs.namespaceMap.MustGet(ns)
		os.MkdirAll(baseDir, os.FileMode(0755))

		func() {
			if symbs.elementMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "element.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.elementMap.Order() {
				symbol := symbs.elementMap.MustGet(name)
				genCtx.ParseSchemaElement(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.groupMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "group.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.groupMap.Order() {
				symbol := symbs.groupMap.MustGet(name)
				genCtx.ParseSchemaGroup(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.attributeMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "attribute.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.attributeMap.Order() {
				symbol := symbs.attributeMap.MustGet(name)
				genCtx.ParseSchemaAttribute(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.attributeGroupMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "attributeGroup.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.attributeGroupMap.Order() {
				symbol := symbs.attributeGroupMap.MustGet(name)
				genCtx.ParseSchemaAttributeGroup(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.simpleTypeMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "simpleType.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.simpleTypeMap.Order() {
				symbol := symbs.simpleTypeMap.MustGet(name)
				genCtx.ParseSchemaSimpleType(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.complexTypeMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "complexType.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.complexTypeMap.Order() {
				symbol := symbs.complexTypeMap.MustGet(name)
				genCtx.ParseSchemaComplexType(symbol.Symbol, symbol.FileName)
			}
		}()
	}
}
