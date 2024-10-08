package render

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

type parseContext struct {
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

func newParseContext(gs *GlobalScope) *parseContext {
	return &parseContext{
		gs:               gs,
		defineState:      make(map[any]definedState),
		order:            tplcontainer.NewSequenceMap[any, *TypeSymbolInfo](),
		forwardDelcOrder: tplcontainer.NewSequenceMap[any, *TypeSymbolInfo](),
	}
}

func (ctx *parseContext) prefix() string {
	var s string
	for i := 0; i < ctx.level; i++ {
		s += "    "
	}
	return s
}

func (ctx *parseContext) output(format string, a ...any) {
	if !ctx.verbose {
		return
	}
	fmt.Print(ctx.prefix())
	fmt.Printf(format, a...)
	fmt.Println()
}

func (ctx *parseContext) nextLevel() func() {
	save := ctx.level
	ctx.level++
	return func() {
		ctx.level = save
	}
}

func (ctx *parseContext) define(symbol any) func() {
	switch ctx.defineState[symbol] {
	case walkingDependency:
		// cycle ref
		return nil
	case defined:
		return nil
	}
	ctx.defineState[symbol] = walkingDependency
	return func() {
		ctx.defineState[symbol] = defined
	}
}

func ParseDocElement(ctx *parseContext, elem *etree.Element) {
	ctx.output("parse element: %s:%s", elem.Space, elem.Tag)
	symbol, ok := ctx.gs.GetElement(elem.NamespaceURI(), elem.Tag)
	if !ok {
		panic("invalid element")
	}

	defer ctx.nextLevel()()
	ParseSchemaElement(ctx, symbol.Symbol, symbol.FileName)
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
func ParseSchemaElement(ctx *parseContext, elem *proto.Element, fileName string) {
	var prefix string
	if len(elem.Ref) > 0 {
		p, symbol, ok := ctx.gs.GetElementByRef(elem, fileName)
		if !ok {
			panic("invalid element")
		}
		elem, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}
	lv := ctx.level

	if len(elem.Type) > 0 {
		typePrefix, symbolType, ok := ctx.gs.GetTypeInFile(elem.Type, fileName)
		if !ok {
			panic("type not found")
		}
		if len(typePrefix) > 0 {
			typePrefix += ":"
		}

		switch typ := symbolType.Symbol.(type) {
		case *proto.ComplexType:
			ctx.output("element: %s%s(%s%s) // %s", prefix, elem.Name, typePrefix, typ.Name, parseAnnotaion(elem.Annotation))

			defer ctx.nextLevel()()
			ParseSchemaComplexType(ctx, typ, symbolType.FileName)

		case *proto.SimpleType:
			ctx.output("element: %s%s(%s%s) // %s", prefix, elem.Name, typePrefix, typ.Name, parseAnnotaion(elem.Annotation))

			defer ctx.nextLevel()()
			ParseSchemaSimpleType(ctx, typ, symbolType.FileName)
		}
	} else if elem.ComplexType != nil {
		ctx.output("element: %s(<complexType>) // %s", elem.Name, parseAnnotaion(elem.Annotation))
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaComplexType(ctx, elem.ComplexType, fileName)
	}

	if elem.Unique != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaUnique(ctx, elem.Unique, fileName)
	}
}

// (annotation?,(simpleContent|complexContent|((group|all|choice|sequence)?,((attribute|attributeGroup)*,anyAttribute?))))
func ParseSchemaComplexType(ctx *parseContext, ct *proto.ComplexType, fileName string) {
	if !ctx.parseRecursive {
		if (ctx.parsingType || ctx.level > 0) && len(ct.Name) > 0 {
			return
		}
		ctx.parsingType = true
		defer func() {
			ctx.parsingType = false
		}()
	}

	switch ctx.defineState[ct] {
	case walkingDependency:
		// cycle ref
		if _, ok := ctx.forwardDelcOrder.Get(ct.Name); !ok {
			ctx.forwardDelcOrder.Set(ct, &TypeSymbolInfo{
				FileName: fileName,
				Define:   &boot.TypeDefineBlock{},
			})
		}
		return
	case defined:
		return
	}
	ctx.defineState[ct] = walkingDependency
	defer func() {
		ctx.order.Set(ct, &TypeSymbolInfo{
			FileName: fileName,
			Define:   &boot.TypeDefineBlock{},
		})
		ctx.defineState[ct] = defined
	}()

	ctx.output("complexType: %s // %s", ct.Name, parseAnnotaion(ct.Annotation))
	lv := ctx.level

	if ct.SimpleContent != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSimpleContent(ctx, ct.SimpleContent, fileName)
	}

	if ct.ComplexContent != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaComplexContent(ctx, ct.ComplexContent, fileName)
	}

	if ct.Group != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaGroup(ctx, ct.Group, fileName)
	}

	if ct.All != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaAll(ctx, ct.All, fileName)
	}

	if ct.Choice != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaChoice(ctx, ct.Choice, fileName)
	}

	if ct.Sequence != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSequence(ctx, ct.Sequence, fileName)
	}

	if len(ct.AttributeList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, attr := range ct.AttributeList {
			ParseSchemaAttribute(ctx, attr, fileName)
		}
	}

	if len(ct.AttributeGroupList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, attrGrp := range ct.AttributeGroupList {
			ParseSchemaAttributeGroup(ctx, attrGrp, fileName)
		}
	}
}

// (annotation?,(restriction|list|union))
func ParseSchemaSimpleType(ctx *parseContext, st *proto.SimpleType, fileName string) {
	if !ctx.parseRecursive {
		if (ctx.parsingType || ctx.level > 0) && len(st.Name) > 0 {
			return
		}
		ctx.parsingType = true
		defer func() {
			ctx.parsingType = false
		}()
	}

	switch ctx.defineState[st] {
	case walkingDependency:
		// cycle ref
		if _, ok := ctx.forwardDelcOrder.Get(st.Name); !ok {
			ctx.forwardDelcOrder.Set(st, &TypeSymbolInfo{
				FileName: fileName,
				Define:   &boot.TypeDefineBlock{},
			})
		}
		return
	case defined:
		return
	}
	ctx.defineState[st] = walkingDependency
	defer func() {
		ctx.order.Set(st, &TypeSymbolInfo{
			FileName: fileName,
			Define:   &boot.TypeDefineBlock{},
		})
		ctx.defineState[st] = defined
	}()

	ctx.output("simpleType: %s // %s", st.Name, parseAnnotaion(st.Annotation))
	lv := ctx.level

	if st.Restriction != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaRestriction(ctx, st.Restriction, fileName)
	}

	if st.List != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaList(ctx, st.List, fileName)
	}

	if st.Union != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaUnion(ctx, st.Union, fileName)
	}
}

// (annotation?,(simpleType?))
func ParseSchemaAttribute(ctx *parseContext, attr *proto.Attribute, fileName string) {
	var prefix string
	if len(attr.Ref) > 0 {
		p, symbol, ok := ctx.gs.GetAttributeByRef(attr, fileName)
		if !ok {
			panic("invalid attribute")
		}
		attr, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}

	if len(attr.Type) > 0 {
		typePrefix, symbolType, ok := ctx.gs.GetSimpleTypeInFile(attr.Type, fileName)
		if !ok {
			panic("simpleType not found")
		}

		ctx.output("attribute: %s%s(%s%s) // %s", prefix, attr.Name, typePrefix, symbolType.Symbol.Name, parseAnnotaion(attr.Annotation))
		defer ctx.nextLevel()()
		ParseSchemaSimpleType(ctx, symbolType.Symbol, symbolType.FileName)
		return
	}

	if attr.SimpleType != nil {
		ctx.output("attribute: %s(<simpleType>) // %s", attr.Name, parseAnnotaion(attr.Annotation))
		defer ctx.nextLevel()()
		ParseSchemaSimpleType(ctx, attr.SimpleType, fileName)
	}
}

// (annotation?),((attribute|attributeGroup)*,anyAttribute?))
func ParseSchemaAttributeGroup(ctx *parseContext, attrGrp *proto.AttributeGroup, fileName string) {
	var prefix string
	if len(attrGrp.Ref) > 0 {
		p, symbol, ok := ctx.gs.GetAttributeGroupByRef(attrGrp, fileName)
		if !ok {
			panic("invalid attributeGroup")
		}
		attrGrp, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}
	ctx.output("attributeGroup: %s%s // %s", prefix, attrGrp.Name, parseAnnotaion(attrGrp.Annotation))

	defer ctx.nextLevel()()
	for _, attr := range attrGrp.AttributeList {
		ParseSchemaAttribute(ctx, attr, fileName)
	}
	for _, attrGrp := range attrGrp.AttributeGroupList {
		ParseSchemaAttributeGroup(ctx, attrGrp, fileName)
	}
}

// (annotation?,(element|group|choice|sequence|any)*)
func ParseSchemaSequence(ctx *parseContext, seq *proto.Sequence, fileName string) {
	ctx.output("sequence // %s", parseAnnotaion(seq.Annotation))
	lv := ctx.level

	if len(seq.ElementList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, elem := range seq.ElementList {
			ParseSchemaElement(ctx, elem, fileName)
		}
	}

	if len(seq.GroupList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, grp := range seq.GroupList {
			ParseSchemaGroup(ctx, grp, fileName)
		}
	}

	if len(seq.ChoiceList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, ch := range seq.ChoiceList {
			ParseSchemaChoice(ctx, ch, fileName)
		}
	}

	if seq.Sequence != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSequence(ctx, seq.Sequence, fileName)
	}

	if len(seq.AnyList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, an := range seq.AnyList {
			ParseSchemaAny(ctx, an, fileName)
		}
	}
}

// (annotation?,(all|choice|sequence)?)
func ParseSchemaGroup(ctx *parseContext, grp *proto.Group, fileName string) {
	var prefix string
	if len(grp.Ref) > 0 {
		p, symbol, ok := ctx.gs.GetGroupByRef(grp, fileName)
		if !ok {
			panic("invalid attributeGroup")
		}
		grp, fileName = symbol.Symbol, symbol.FileName
		if len(p) > 0 {
			prefix = p + ":"
		}
	}
	ctx.output("group: %s%s // %s", prefix, grp.Name, parseAnnotaion(grp.Annotation))
	lv := ctx.level

	if grp.Choice != nil {
		defer ctx.nextLevel()()
		ParseSchemaChoice(ctx, grp.Choice, fileName)
	}

	if grp.Sequence != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSequence(ctx, grp.Sequence, fileName)
	}
}

// (annotation?,(element|group|choice|sequence|any)*)
func ParseSchemaChoice(ctx *parseContext, ch *proto.Choice, fileName string) {
	ctx.output("choice // %s", parseAnnotaion(ch.Annotation))
	lv := ctx.level

	if len(ch.ElementList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, elem := range ch.ElementList {
			ParseSchemaElement(ctx, elem, fileName)
		}
	}

	if len(ch.GroupList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, grp := range ch.GroupList {
			ParseSchemaGroup(ctx, grp, fileName)
		}
	}
	if ch.Choice != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaChoice(ctx, ch.Choice, fileName)
	}

	if ch.Sequence != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSequence(ctx, ch.Sequence, fileName)
	}

	if ch.Any != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaAny(ctx, ch.Any, fileName)
	}
}

// (annotation?,element*)
func ParseSchemaAll(ctx *parseContext, all *proto.All, fileName string) {
	ctx.output("all")

	if len(all.ElementList) > 0 {
		defer ctx.nextLevel()()
		for _, elem := range all.ElementList {
			ParseSchemaElement(ctx, elem, fileName)
		}
	}
}

// (annotation?,(restriction|extension))
func ParseSchemaSimpleContent(ctx *parseContext, sc *proto.SimpleContent, fileName string) {
	ctx.output("simpleContent")

	if sc.Extension != nil {
		defer ctx.nextLevel()()
		ParseSchemaExtension(ctx, sc.Extension, fileName)
	}
}

// (annotation?,(restriction|extension))
func ParseSchemaComplexContent(ctx *parseContext, cc *proto.ComplexContent, fileName string) {
	ctx.output("complexContent")
	lv := ctx.level

	if cc.Restriction != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaRestriction(ctx, cc.Restriction, fileName)
	}

	if cc.Extension != nil {
		defer ctx.nextLevel()()
		ParseSchemaExtension(ctx, cc.Extension, fileName)
	}
}

// (annotation?)
func ParseSchemaAny(ctx *parseContext, an *proto.Any, fileName string) {
	ctx.output("any")
}

// (annotation?,((group|all|choice|sequence)?,((attribute|attributeGroup)*,anyAttribute?)))
func ParseSchemaExtension(ctx *parseContext, ext *proto.Extension, fileName string) {
	if len(ext.Base) == 0 {
		panic("invalid extension")
	}
	ctx.output("extension: %s", ext.Base)
	lv := ctx.level

	_, symbolType, ok := ctx.gs.GetTypeInFile(ext.Base, fileName)
	if !ok {
		panic("type not found")
	}
	switch typ := symbolType.Symbol.(type) {
	case *proto.ComplexType:
		ParseSchemaComplexType(ctx, typ, symbolType.FileName)

	case *proto.SimpleType:

		ParseSchemaSimpleType(ctx, typ, symbolType.FileName)
	}

	if ext.Sequence != nil {
		defer ctx.nextLevel()()
		ParseSchemaSequence(ctx, ext.Sequence, fileName)
	}

	if ext.Choice != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaChoice(ctx, ext.Choice, fileName)
	}

	if ext.Group != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaGroup(ctx, ext.Group, fileName)
	}

	if len(ext.AttributeList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, attr := range ext.AttributeList {
			ParseSchemaAttribute(ctx, attr, fileName)
		}
	}

	if len(ext.AttributeGroupList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, attrGrp := range ext.AttributeGroupList {
			ParseSchemaAttributeGroup(ctx, attrGrp, fileName)
		}
	}
}

// simpleType: (annotation?,(simpleType?,(minExclusive|minInclusive|maxExclusive|maxInclusive|totalDigits|fractionDigits|length|minLength|maxLength|enumeration|whiteSpace|pattern)*))
// simpleContent: (annotation?,(simpleType?,(minExclusive |minInclusive|maxExclusive|maxInclusive|totalDigits|fractionDigits|length|minLength|maxLength|enumeration|whiteSpace|pattern)*)?,((attribute|attributeGroup)*,anyAttribute?))
// complexContent: (annotation?,(group|all|choice|sequence)?,((attribute|attributeGroup)*,anyAttribute?))
func ParseSchemaRestriction(ctx *parseContext, rest *proto.Restriction, fileName string) {
	lv := ctx.level
	if len(rest.Base) > 0 {
		ctx.output("restriction: %s // %s", rest.Base, parseAnnotaion(rest.Annotation))
	} else if rest.SimpleType != nil {
		ctx.output("restriction: <simpleType> // %s", rest.Base, parseAnnotaion(rest.Annotation))
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSimpleType(ctx, rest.SimpleType, fileName)
	} else {
		panic("invalid restriction")
	}

	if len(rest.EnumerationList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, enum := range rest.EnumerationList {
			ParseSchemaEnumeration(ctx, enum, fileName)
		}
	}

	if rest.Group != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaGroup(ctx, rest.Group, fileName)
	}

	if rest.Sequence != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSequence(ctx, rest.Sequence, fileName)
	}

	if len(rest.AttributeList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, attr := range rest.AttributeList {
			ParseSchemaAttribute(ctx, attr, fileName)
		}
	}
}

func ParseSchemaEnumeration(ctx *parseContext, enum *proto.Enumeration, fileName string) {
	ctx.output("enumeration: %s // %s", enum.Value, parseAnnotaion(enum.Annotation))
}

// (annotation?,(simpleType*))
func ParseSchemaUnion(ctx *parseContext, uni *proto.Union, fileName string) {
	ctx.output("union")
	lv := ctx.level

	if len(uni.MemberTypes) > 0 {
		sli := strings.Split(uni.MemberTypes, " ")
		if len(sli) == 0 {
			panic("empty memberTypes")
		}

		defer ctx.nextLevel()()
		for _, typ := range sli {
			_, symbolType, ok := ctx.gs.GetTypeInFile(typ, fileName)
			if !ok {
				panic("type not found")
			}
			switch typ := symbolType.Symbol.(type) {
			case *proto.ComplexType:
				ParseSchemaComplexType(ctx, typ, symbolType.FileName)

			case *proto.SimpleType:
				ParseSchemaSimpleType(ctx, typ, symbolType.FileName)
			}
		}
	}

	if len(uni.SimpleTypeList) > 0 {
		ctx.level = lv
		defer ctx.nextLevel()()
		for _, st := range uni.SimpleTypeList {
			ParseSchemaSimpleType(ctx, st, fileName)
		}
	}
}

// (annotation?,(simpleType?))
func ParseSchemaList(ctx *parseContext, lst *proto.List, fileName string) {
	ctx.output("list")
	lv := ctx.level

	if len(lst.ItemType) > 0 {
		_, symbolType, ok := ctx.gs.GetTypeInFile(lst.ItemType, fileName)
		if !ok {
			panic("type not found")
		}
		switch typ := symbolType.Symbol.(type) {
		case *proto.ComplexType:
			ParseSchemaComplexType(ctx, typ, symbolType.FileName)

		case *proto.SimpleType:
			ParseSchemaSimpleType(ctx, typ, symbolType.FileName)
		}
	}

	if lst.SimpleType != nil {
		ctx.level = lv
		defer ctx.nextLevel()()
		ParseSchemaSimpleType(ctx, lst.SimpleType, fileName)
	}
}

// (annotation?,(selector,field+))
func ParseSchemaUnique(ctx *parseContext, uniq *proto.Unique, fileName string) {
	ctx.output("unique")

	if uniq.Selector != nil {
		ctx.output("selector: %s", uniq.Selector.Xpath)
	}

	if uniq.Field != nil {
		ctx.output("field: %s", uniq.Field.Xpath)
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

	// ctx := newParseContext(gs)
	// ParseSchemaElement(ctx, symbol.Symbol, symbol.FileName)

	zr, err := zip.OpenReader("data/test.pptx")
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
		if file.Name == "ppt/slides/slide1.xml" {
			TestParse(fp, gs)
		}
	}

	fmt.Println("exit.")
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

	orderCtx := newParseContext(gs)
	orderCtx.parseRecursive = true
	orderCtx.verbose = false
	ParseDocElement(orderCtx, root)

	genCtx := newParseContext(gs)
	genCtx.parseRecursive = false
	genCtx.verbose = true
	for _, symb := range orderCtx.order.Order() {
		info := orderCtx.order.MustGet(symb)
		switch typ := symb.(type) {
		case *proto.ComplexType:
			ParseSchemaComplexType(genCtx, typ, info.FileName)
		case *proto.SimpleType:
			ParseSchemaSimpleType(genCtx, typ, info.FileName)
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
	// 		ParseSchemaComplexType(genCtx, typ, info.FileName)
	// 	case *proto.SimpleType:
	// 		ParseSchemaSimpleType(genCtx, typ, info.FileName)
	// 	}
	// }

	// 按命名空间生成
	genCtx := newParseContext(gs)
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
				ParseSchemaElement(genCtx, symbol.Symbol, symbol.FileName)
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
				ParseSchemaGroup(genCtx, symbol.Symbol, symbol.FileName)
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
				ParseSchemaAttribute(genCtx, symbol.Symbol, symbol.FileName)
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
				ParseSchemaAttributeGroup(genCtx, symbol.Symbol, symbol.FileName)
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
				ParseSchemaSimpleType(genCtx, symbol.Symbol, symbol.FileName)
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
				ParseSchemaComplexType(genCtx, symbol.Symbol, symbol.FileName)
			}
		}()
	}
}
