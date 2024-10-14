package render

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/tutils/xmlschema/data"
	"github.com/tutils/xmlschema/proto"
	"github.com/tutils/xmlschema/tplcontainer"
)

const (
	// namespace
	XMLNamespace = "http://www.w3.org/XML/1998/namespace"
	XSNamespace  = "http://www.w3.org/2001/XMLSchema"

	// prefix
	XMLNamespacePrefix = "xml"

	// schema location
	XMLSchemaLocation       = "http://www.w3.org/2001/xml.xsd"
	XML200103SchemaLocation = "http://www.w3.org/2001/03/xml.xsd"
	XSSchemaLocation        = "http://www.w3.org/2001/XMLSchema.xsd"

	DCSchemaLocation       = "http://dublincore.org/schemas/xmls/qdc/2003/04/02/dc.xsd"
	DCTermsSchemaLocation  = "http://dublincore.org/schemas/xmls/qdc/2003/04/02/dcterms.xsd"
	DCMITypeSchemaLocation = "http://dublincore.org/schemas/xmls/qdc/2003/04/02/dcmitype.xsd"

	// local schema location
	XMLLocalSchemaLocation       = "embed:///schemas/xml/xml.xsd"
	XML200103LocalSchemaLocation = "embed:///schemas/xml/xml200103.xsd"
	XSLocalSchemaLocation        = "embed:///schemas/xml/XMLSchema.xsd"

	DCLocalSchemaLocation       = "embed:///schemas/dc/dc.xsd"
	DCTermsLocalSchemaLocation  = "embed:///schemas/dc/dcterms.xsd"
	DCMITypeLocalSchemaLocation = "embed:///schemas/dc/dcmitype.xsd"
)

var (
	remote2local = map[string]string{
		XMLSchemaLocation:       XMLLocalSchemaLocation,
		XML200103SchemaLocation: XML200103LocalSchemaLocation,
		XSSchemaLocation:        XSLocalSchemaLocation,
		DCSchemaLocation:        DCLocalSchemaLocation,
		DCTermsSchemaLocation:   DCTermsLocalSchemaLocation,
		DCMITypeSchemaLocation:  DCMITypeLocalSchemaLocation,
	}

	xsFs        *FileScope
	xmlFs       *FileScope
	xml200103Fs *FileScope

	fakeAnySimpleType = &proto.SimpleType{
		XMLName: xml.Name{
			Space: XSNamespace,
			Local: "simpleType",
		},
		Name: "anySimpleType",
		Restriction: &proto.Restriction{
			XMLName: xml.Name{
				Space: XSNamespace,
				Local: "restriction",
			},
			Base: "xs:anySimpleType",
		},
	}
)

func init() {
	cache := make(FileCache)

	xsFs = newFileScope()
	xsFs.loadSchema(XSSchemaLocation, cache)

	xmlFs = newFileScope()
	xmlFs.loadSchema(XMLSchemaLocation, cache)

	xml200103Fs = newFileScope()
	xml200103Fs.loadSchema(XML200103SchemaLocation, cache)
}

type FileCache = map[string]*proto.Schema

type FileScope struct {
	name      string
	namespace string

	// [prefix, namespace]
	prefixMap *tplcontainer.SequenceMap[string, string]

	schema *proto.Schema

	simpleTypeList     []string
	complexTypeList    []string
	elementList        []string
	groupList          []string
	attributeList      []string
	attributeGroupList []string

	// [namespace, SymbolMap], namespaceMap 用于校验引用是否合法
	namespaceMap *tplcontainer.SequenceMap[string, *SymbolMap]

	// [file, record]
	cache FileCache
}

func newFileScope() *FileScope {
	return &FileScope{
		prefixMap:    tplcontainer.NewSequenceMap[string, string](),
		namespaceMap: tplcontainer.NewSequenceMap[string, *SymbolMap](),
	}
}

func (fs *FileScope) loadSchema(name string, cache FileCache) {
	if cache == nil {
		cache = FileCache{
			XMLSchemaLocation:       xmlFs.schema,
			XML200103SchemaLocation: xml200103Fs.schema,
			XSSchemaLocation:        xsFs.schema,
		}
	}
	fs.cache = cache
	defer func() {
		fs.cache = nil
	}()

	loadRecord := make(map[string]loadState)
	fs.name = name
	if name != XSSchemaLocation {
		// 预先加载XMLSchema
		fs.buildNamespaceMap(XSSchemaLocation, loadRecord)
	}
	fs.schema = fs.buildNamespaceMap(name, loadRecord)

	fs.namespace = fs.schema.TargetNamespace
	fs.buildPrefixMap()

	// 校验引用
	if !fs.verify() {
		panic("invalid schema")
	}

	fs.addSymbols()
}

func (fs *FileScope) addSymbols() {
	schema := fs.schema

	fs.complexTypeList = make([]string, len(schema.ComplexTypeList))
	for i, ct := range schema.ComplexTypeList {
		fs.complexTypeList[i] = ct.Name
	}

	fs.simpleTypeList = make([]string, len(schema.SimpleTypeList))
	for i, st := range schema.SimpleTypeList {
		fs.simpleTypeList[i] = st.Name
	}

	fs.elementList = make([]string, len(schema.ElementList))
	for i, elem := range schema.ElementList {
		fs.elementList[i] = elem.Name
	}

	fs.groupList = make([]string, len(schema.GroupList))
	for i, grp := range schema.GroupList {
		fs.groupList[i] = grp.Name
	}

	fs.attributeList = make([]string, len(schema.AttributeList))
	for i, attr := range schema.AttributeList {
		fs.attributeList[i] = attr.Name
	}

	fs.attributeGroupList = make([]string, len(schema.AttributeGroupList))
	for i, attrGrp := range schema.AttributeGroupList {
		fs.attributeGroupList[i] = attrGrp.Name
	}
}

// buildPrefixMap 构建前缀到命名空间的映射
// 如："p" -> "http://schemas.openxmlformats.org/presentationml/2006/main"
func (fs *FileScope) buildPrefixMap() {
	schema := fs.schema

	// 添加xml命名空间映射
	fs.prefixMap.Set(XMLNamespacePrefix, XMLNamespace)

	if len(schema.Xmlns) > 0 {
		// 存在默认命名空间，即不带前缀时的命名空间。在引用schema中的定义时使用
		fs.prefixMap.Set("", schema.Xmlns)
	}

	// 遍历命名空间
	for _, xmlns := range schema.XmlnsList {
		fs.prefixMap.Set(xmlns.Name.Local, xmlns.Value)
	}

	// 从ImportList更新prefixMap
	for _, imp := range schema.ImportList {
		if len(imp.Namespace) == 0 {
			panic("empty namespace")
		}
	}
}

// parseSchema 只解析当前文件，不包含import和include
func (fs *FileScope) parseSchema(name string) *proto.Schema {
	// dir := filepath.Dir(name)
	// name = filepath.Join(dir, filepath.Base(name))

	if rec, ok := fs.cache[name]; ok {
		fmt.Println("loading schema (cache):", name)
		return rec
	}
	fmt.Println("loading schema:", name)

	var newName string
	if local, ok := remote2local[name]; ok {
		newName = local
	} else {
		newName = name
	}

	u, err := url.Parse(newName)
	if err != nil {
		panic(err)
	}

	var fp io.ReadCloser
	switch u.Scheme {
	case "embed":
		fp, err = data.Content.Open(u.Path[1:])
	case "file", "":
		fp, err = os.Open(u.Path)
	default:
		err = errors.New("unsupported scheme")
	}
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	dec := xml.NewDecoder(fp)
	schema := &proto.Schema{}
	if err := dec.Decode(schema); err != nil {
		panic(err)
	}

	fs.cache[name] = schema

	return schema
}

// buildNamespaceMap 递归遍历schema定义文件，构建上下文符号表，返回根节点的原始schema数据
func (fs *FileScope) buildNamespaceMap(name string, loadRecord map[string]loadState) *proto.Schema {
	if loadRecord == nil {
		loadRecord = make(map[string]loadState)
	} else {
		switch loadRecord[name] {
		case loading:
			fmt.Println("cycle xsd ref:", name)
			return fs.cache[name]
		case loaded:
			return fs.cache[name]
		}
	}
	defer func() {
		loadRecord[name] = loaded
	}()
	loadRecord[name] = loading

	// 解析当前节点，将当前文件的符号表（不含外部文件的符号），添加到上下文符号表（含有外部符号）
	schema := fs.parseSchema(name)
	ns := schema.TargetNamespace
	symbs, ok := fs.namespaceMap.Get(ns)
	if !ok {
		symbs = NewSymbolMap()
		fs.namespaceMap.Set(ns, symbs)
	}
	symbs.AddSymbolsFromSchema(schema, name)
	if name == XSSchemaLocation {
		// add build-in ur-type
		symbs.addSimpleType(fakeAnySimpleType, name)
	}

	var dir string
	var local bool
	u, err := url.Parse(name)
	if err != nil {
		panic(err)
	}
	if len(u.Scheme) == 0 {
		local = true
		dir = filepath.Dir(name)
	} else {
		local = false
		dir = u.Scheme + "://" + u.Host + path.Dir(u.Path)
	}

	// 从ImportList递归遍历schema定义文件
	for _, imp := range schema.ImportList {
		if len(imp.Namespace) == 0 {
			panic("empty namespace")
		}

		if len(imp.SchemaLocation) > 0 {
			var impFile string
			u, err := url.Parse(imp.SchemaLocation)
			if err != nil {
				panic(err)
			}
			if len(u.Scheme) == 0 {
				if local {
					// 文件相对路径
					impFile = filepath.Join(dir, imp.SchemaLocation)
				} else {
					// URI相对路径
					impFile, err = url.JoinPath(dir, imp.SchemaLocation)
					if err != nil {
						panic(err)
					}
				}
			} else {
				impFile = imp.SchemaLocation
			}
			fs.buildNamespaceMap(impFile, loadRecord)
		}
	}

	// 从IncludeList递归遍历schema定义文件
	for _, inc := range schema.IncludeList {
		if len(inc.SchemaLocation) > 0 {
			incFile := filepath.Join(dir, inc.SchemaLocation)
			fs.buildNamespaceMap(incFile, loadRecord)
		}
	}

	return schema
}

func (fs *FileScope) verifyComplexType(ct *proto.ComplexType) bool {
	// simpleContent | complexContent | (attribute, attributeGroup)
	hasCheck := false
	if len(ct.AttributeList) > 0 {
		for _, attr := range ct.AttributeList {
			if !fs.verifyAttribute(attr) {
				return false
			}
		}
		hasCheck = true
	}
	if len(ct.AttributeGroupList) > 0 {
		for _, attrGrp := range ct.AttributeGroupList {
			if !fs.verifyAttributeGroup(attrGrp) {
				return false
			}
		}
		hasCheck = true
	}
	if ct.Sequence != nil {
		if !fs.verifySequence(ct.Sequence) {
			return false
		}
		hasCheck = true
	}
	if ct.Choice != nil {
		if !fs.verifyChoice(ct.Choice) {
			return false
		}
		hasCheck = true
	}
	if ct.Group != nil {
		if !fs.verifyGroup(ct.Group) {
			return false
		}
		hasCheck = true
	}
	if ct.All != nil {
		if !fs.verifyAll(ct.All) {
			return false
		}
		hasCheck = true
	}

	if ct.SimpleContent != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifySimpleContent(ct.SimpleContent) {
			return false
		}
		hasCheck = true
	}

	if ct.ComplexContent != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyComplexContent(ct.ComplexContent) {
			return false
		}
		hasCheck = true
	}

	// 允许空类型定义

	return true
}

func (fs *FileScope) verifySimpleType(st *proto.SimpleType) bool {
	// Restriction | Union | List
	hasCheck := false
	if st.Union != nil {
		if !fs.verifyUnion(st.Union) {
			return false
		}
		hasCheck = true
	}

	if st.List != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyList(st.List) {
			return false
		}
		hasCheck = true
	}

	if st.Restriction != nil {
		if !fs.verifyRestriction(st.Restriction) {
			return false
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty simpleType define")
}

func (fs *FileScope) verifyElement(elem *proto.Element) bool {
	// Type | Ref
	if len(elem.Type) != 0 && len(elem.Ref) != 0 {
		panic("multi attribute")
	}

	if len(elem.Name) > 0 {
		hasCheck := false
		if len(elem.Type) > 0 {
			if !fs.verifyTypeRef(elem.Type) {
				return false
			}
			hasCheck = true

		}

		if elem.ComplexType != nil {
			if hasCheck {
				panic("multi child element")
			}
			if !fs.verifyComplexType(elem.ComplexType) {
				return false
			}
			hasCheck = true
		}

		// if hasCheck {
		// 	return true
		// }

		// panic("empty element define")
		return true
	}

	if len(elem.Ref) > 0 {
		return fs.verifyElementRef(elem.Ref)
	}

	panic("empty element")
}

func (fs *FileScope) verifyAttribute(attr *proto.Attribute) bool {
	// Type | Ref
	if len(attr.Name) > 0 {
		hasCheck := false
		if len(attr.Type) > 0 {
			if !fs.verifyTypeRef(attr.Type) {
				return false
			}
			hasCheck = true
		}

		if attr.SimpleType != nil {
			if hasCheck {
				panic("multi child element")
			}
			if !fs.verifySimpleType(attr.SimpleType) {
				return false
			}
			hasCheck = true
		}

		if hasCheck {
			return true
		}

		if attr.Name == "value" {
			return true
		}

		panic("empty attribute define")
	}

	if len(attr.Ref) > 0 {
		return fs.verifyAttributeRef(attr.Ref)
	}

	panic("empty attribute")
}

func (fs *FileScope) verifyGroup(grp *proto.Group) bool {
	if len(grp.Name) != 0 && len(grp.Ref) != 0 {
		panic("multi attribute")
	}

	if len(grp.Name) > 0 {
		// Group Define
		// Choice | Sequence
		hasCheck := false
		if grp.Choice != nil {
			if !fs.verifyChoice(grp.Choice) {
				return false
			}
			hasCheck = true
		}

		if grp.Sequence != nil {
			if hasCheck {
				panic("multi attribute")
			}
			if !fs.verifySequence(grp.Sequence) {
				return false
			}
			hasCheck = true
		}

		if hasCheck {
			return true
		}

		panic("empty group define")
	}

	if len(grp.Ref) > 0 {
		// Group Ref
		return fs.verifyGroupRef(grp.Ref)
	}

	panic("empty group")
}

func (fs *FileScope) verifyAttributeGroup(attrGrp *proto.AttributeGroup) bool {
	if len(attrGrp.Name) != 0 && len(attrGrp.Ref) != 0 {
		panic("multi attribute")
	}

	if len(attrGrp.Name) > 0 {
		// AttributeGroup Define
		for _, attr := range attrGrp.AttributeList {
			if !fs.verifyAttribute(attr) {
				return false
			}
		}
		for _, nestAttrGrp := range attrGrp.AttributeGroupList {
			if !fs.verifyAttributeGroup(nestAttrGrp) {
				return false
			}
		}

		// 允许空定义
		return true
	}

	if len(attrGrp.Ref) > 0 {
		// AttributeGroup Ref
		return fs.verifyAttributeGroupRef(attrGrp.Ref)
	}

	panic("empty attributeGroup define")
}

func (fs *FileScope) verifyChoice(ch *proto.Choice) bool {
	hasCheck := false
	if ch.Sequence != nil {
		if !fs.verifySequence(ch.Sequence) {
			return false
		}
		hasCheck = true
	}

	if ch.Choice != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyChoice(ch.Choice) {
			return false
		}
		hasCheck = true
	}

	if ch.Any != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyAny(ch.Any) {
			return false
		}
		hasCheck = true
	}

	if len(ch.ElementList) > 0 {
		for _, elem := range ch.ElementList {
			if !fs.verifyElement(elem) {
				return false
			}
		}
		hasCheck = true
	}

	if len(ch.GroupList) > 0 {
		for _, grp := range ch.GroupList {
			if !fs.verifyGroup(grp) {
				return false
			}
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty choice")
}

// 只有XMLSchema.xsd中的sequence存在Annotation
// func (fs *FileScope) verifySequence_old(seq *proto.Sequence) bool {
// 	hasCheck := false
// 	if len(seq.ChoiceList) > 0 {
// 		for _, ch := range seq.ChoiceList {
// 			if !fs.verifyChoice(ch) {
// 				return false
// 			}
// 		}
// 		hasCheck = true
// 	}

// 	if len(seq.AnyList) > 0 {
// 		if hasCheck {
// 			panic("multi child element")
// 		}
// 		hasCheck = true
// 	}

// 	if len(seq.ElementList) > 0 {
// 		for _, elem := range seq.ElementList {
// 			if !fs.verifyElement(elem) {
// 				return false
// 			}
// 		}
// 		hasCheck = true
// 	}

// 	if len(seq.GroupList) > 0 {
// 		for _, grp := range seq.GroupList {
// 			if !fs.verifyGroup(grp) {
// 				return false
// 			}
// 		}
// 		hasCheck = true
// 	}

// 	return true
// }

// 只有XMLSchema.xsd中的sequence存在Annotation
func (fs *FileScope) verifySequence(seq *proto.Sequence) bool {
	for _, part := range seq.NestedParticleList {
		switch ch := part.(type) {
		case *proto.Element:
			if !fs.verifyElement(ch) {
				return false
			}
		case *proto.Group:
			if !fs.verifyGroup(ch) {
				return false
			}
		case *proto.Choice:
			if !fs.verifyChoice(ch) {
				return false
			}
		case *proto.Sequence:
			if !fs.verifySequence(ch) {
				return false
			}
		case *proto.Any:
			if !fs.verifyAny(ch) {
				return false
			}
		}
	}

	return true
}

func (fs *FileScope) verifyAny(an *proto.Any) bool {
	return true
}

func (fs *FileScope) verifyUnion(uni *proto.Union) bool {
	hasCheck := false
	if len(uni.MemberTypes) > 0 {
		sli := strings.Split(uni.MemberTypes, " ")
		if len(sli) == 0 {
			panic("empty memberTypes")
		}
		for _, typ := range sli {
			if !fs.verifyTypeRef(typ) {
				return false
			}
		}
		hasCheck = true
	}

	if len(uni.SimpleTypeList) > 0 {
		for _, st := range uni.SimpleTypeList {
			if !fs.verifySimpleType(st) {
				return false
			}
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty union define")
}

func (fs *FileScope) verifyList(lst *proto.List) bool {
	hasCheck := false
	if len(lst.ItemType) > 0 {
		if !fs.verifyTypeRef(lst.ItemType) {
			return false
		}
		hasCheck = true
	}

	if lst.SimpleType != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifySimpleType(lst.SimpleType) {
			return false
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty list define")
}

func (fs *FileScope) verifyAll(all *proto.All) bool {
	if len(all.ElementList) == 0 {
		panic("empty element")
	}

	for _, elem := range all.ElementList {
		if !fs.verifyElement(elem) {
			return false
		}
	}
	return true
}

func (fs *FileScope) verifySimpleContent(sc *proto.SimpleContent) bool {
	hasCheck := false
	if sc.Extension != nil {
		if !fs.verifyExtension(sc.Extension) {
			return false
		}
		hasCheck = true
	}

	if sc.Restriction != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyRestriction(sc.Restriction) {
			return false
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty simpleContent")
}

func (fs *FileScope) verifyComplexContent(cc *proto.ComplexContent) bool {
	hasCheck := false
	if cc.Extension != nil {
		if !fs.verifyExtension(cc.Extension) {
			return false
		}
		hasCheck = true
	}

	if cc.Restriction != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyRestriction(cc.Restriction) {
			return false
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty complexContent")
}

func (fs *FileScope) verifyExtension(ext *proto.Extension) bool {
	hasCheck := false
	if ext.Sequence != nil {
		if !fs.verifySequence(ext.Sequence) {
			return false
		}
		hasCheck = true
	}

	if ext.Choice != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyChoice(ext.Choice) {
			return false
		}
		hasCheck = true
	}

	if ext.Group != nil {
		if hasCheck {
			panic("multi child element")
		}
		if !fs.verifyGroup(ext.Group) {
			return false
		}
		hasCheck = true
	}

	if len(ext.Base) > 0 {
		if !fs.verifyTypeRef(ext.Base) {
			return false
		}
		hasCheck = true
	}

	if len(ext.AttributeList) > 0 {
		for _, attr := range ext.AttributeList {
			if !fs.verifyAttribute(attr) {
				return false
			}
		}
		hasCheck = true
	}

	if len(ext.AttributeGroupList) > 0 {
		for _, attrGrp := range ext.AttributeGroupList {
			if !fs.verifyAttributeGroup(attrGrp) {
				return false
			}
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty extension")
}

func (fs *FileScope) verifyRestriction(rest *proto.Restriction) bool {
	hasCheck := false
	if len(rest.Base) > 0 {
		if !fs.verifyTypeRef(rest.Base) {
			return false
		}
		hasCheck = true
	}

	if rest.SimpleType != nil {
		if !fs.verifySimpleType(rest.SimpleType) {
			return false
		}
		hasCheck = true
	}

	if hasCheck {
		return true
	}

	panic("empty restriction")
}

func PrefixAndLocal(nameWithPrefix string) (string, string) {
	sli := strings.Split(nameWithPrefix, ":")
	var prefix, local string
	switch len(sli) {
	case 1:
		prefix = ""
		local = nameWithPrefix
	case 2:
		prefix = sli[0]
		local = sli[1]
	default:
		panic("invalid name")
	}
	return prefix, local
}

func (fs *FileScope) getSymbolMap(nameWithPrefix string) (string, string, *SymbolMap, bool) {
	prefix, local := PrefixAndLocal(nameWithPrefix)
	ns, ok := fs.prefixMap.Get(prefix)
	if !ok {
		return prefix, local, nil, false
	}

	symbs, ok := fs.namespaceMap.Get(ns)
	return prefix, local, symbs, ok
}

// GetSimpleType 根据名字获取符号的前缀和结构。无论成功失败，前缀都将会正确返回
// 名字应该包含前缀，如果没有前缀，默认从顶层命名空间中寻找符号
func (fs *FileScope) GetSimpleType(nameWithPrefix string) (string, *Symbol[*proto.SimpleType], bool) {
	prefix, local, symbs, ok := fs.getSymbolMap(nameWithPrefix)
	if !ok {
		return prefix, nil, false
	}

	symb, ok := symbs.GetSimpleType(local)
	return prefix, symb, ok
}

// GetComplexType 根据名字获取符号的前缀和结构。无论成功失败，前缀都将会正确返回
// 名字应该包含前缀，如果没有前缀，默认从顶层命名空间中寻找符号
func (fs *FileScope) GetComplexType(nameWithPrefix string) (string, *Symbol[*proto.ComplexType], bool) {
	prefix, local, symbs, ok := fs.getSymbolMap(nameWithPrefix)
	if !ok {
		return prefix, nil, false
	}

	symb, ok := symbs.GetComplexType(local)
	return prefix, symb, ok
}

func (fs *FileScope) GetType(nameWithPrefix string) (string, *Symbol[any], bool) {
	prefix, local, symbs, ok := fs.getSymbolMap(nameWithPrefix)
	if !ok {
		return prefix, nil, false
	}

	ct, ok := symbs.GetComplexType(local)
	if ok {
		return prefix, &Symbol[any]{
			Symbol:   ct.Symbol,
			FileName: ct.FileName,
		}, true
	}

	st, ok := symbs.GetSimpleType(local)
	if ok {
		return prefix, &Symbol[any]{
			Symbol:   st.Symbol,
			FileName: st.FileName,
		}, true
	}

	return prefix, nil, false
}

// GetElement 根据名字获取符号的前缀和结构。无论成功失败，前缀都将会正确返回
// 名字应该包含前缀，如果没有前缀，默认从顶层命名空间中寻找符号
func (fs *FileScope) GetElement(nameWithPrefix string) (string, *Symbol[*proto.Element], bool) {
	prefix, local, symbs, ok := fs.getSymbolMap(nameWithPrefix)
	if !ok {
		return prefix, nil, false
	}

	symb, ok := symbs.GetElement(local)
	return prefix, symb, ok
}

// GetGroup 根据名字获取符号的前缀和结构。无论成功失败，前缀都将会正确返回
// 名字应该包含前缀，如果没有前缀，默认从顶层命名空间中寻找符号
func (fs *FileScope) GetGroup(nameWithPrefix string) (string, *Symbol[*proto.Group], bool) {
	prefix, local, symbs, ok := fs.getSymbolMap(nameWithPrefix)
	if !ok {
		return prefix, nil, false
	}

	symb, ok := symbs.GetGroup(local)
	return prefix, symb, ok
}

// GetAttribute 根据名字获取符号的前缀和结构。无论成功失败，前缀都将会正确返回
// 名字应该包含前缀，如果没有前缀，默认从顶层命名空间中寻找符号
func (fs *FileScope) GetAttribute(nameWithPrefix string) (string, *Symbol[*proto.Attribute], bool) {
	prefix, local, symbs, ok := fs.getSymbolMap(nameWithPrefix)
	if !ok {
		return prefix, nil, false
	}

	symb, ok := symbs.GetAttribute(local)
	return prefix, symb, ok
}

// GetAttributeGroup 根据名字获取符号的前缀和结构。无论成功失败，前缀都将会正确返回
// 名字应该包含前缀，如果没有前缀，默认从顶层命名空间中寻找符号
func (fs *FileScope) GetAttributeGroup(nameWithPrefix string) (string, *Symbol[*proto.AttributeGroup], bool) {
	prefix, local, symbs, ok := fs.getSymbolMap(nameWithPrefix)
	if !ok {
		return prefix, nil, false
	}

	symb, ok := symbs.GetAttributeGroup(local)
	return prefix, symb, ok
}

func (fs *FileScope) verifyTypeRef(name string) bool {
	_, _, ok := fs.GetComplexType(name)
	if ok {
		return true
	}

	_, _, ok = fs.GetSimpleType(name)
	return ok

	// if _, ok := ignorePrefix[prefix]; ok {
	// 	return true
	// }

	// ns := fs.prefixMap.MustGet(prefix)
	// if _, ok := ignoreNamepace[ns]; ok {
	// 	return true
	// }

	// return false
}

func (fs *FileScope) verifyElementRef(name string) bool {
	_, _, ok := fs.GetElement(name)
	return ok
}

func (fs *FileScope) verifyGroupRef(name string) bool {
	_, _, ok := fs.GetGroup(name)
	return ok
}

func (fs *FileScope) verifyAttributeRef(name string) bool {
	_, _, ok := fs.GetAttribute(name)
	return ok
}

func (fs *FileScope) verifyAttributeGroupRef(name string) bool {
	_, _, ok := fs.GetAttributeGroup(name)
	return ok
}

// verify 验证整个Schema的合法性
func (fs *FileScope) verify() bool {
	schema := fs.schema

	for _, attr := range schema.AttributeList {
		// fmt.Println(attr.Name)
		if !fs.verifyAttribute(attr) {
			return false
		}
	}

	for _, ct := range schema.ComplexTypeList {
		// fmt.Println(ct.Name)
		if !fs.verifyComplexType(ct) {
			return false
		}
	}

	for _, st := range schema.SimpleTypeList {
		// fmt.Println(st.Name)
		if !fs.verifySimpleType(st) {
			return false
		}
	}

	for _, elem := range schema.ElementList {
		// fmt.Println(elem.Name)
		if !fs.verifyElement(elem) {
			return false
		}
	}

	for _, grp := range schema.GroupList {
		// fmt.Println(grp.Name)
		if !fs.verifyGroup(grp) {
			return false
		}
	}

	for _, attrGrp := range schema.AttributeGroupList {
		// fmt.Println(attrGrp.Name)
		if !fs.verifyAttributeGroup(attrGrp) {
			return false
		}
	}

	return true
}
