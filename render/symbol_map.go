package render

import (
	"github.com/tutils/xmlschema/proto"
	"github.com/tutils/xmlschema/tplcontainer"
)

// type Symbol = any

type Symbol[SYMBOL any] struct {
	Symbol   SYMBOL
	FileName string
}

type SymbolMap struct {
	simpleTypeMap     *tplcontainer.SequenceMap[string, *Symbol[*proto.SimpleType]]
	complexTypeMap    *tplcontainer.SequenceMap[string, *Symbol[*proto.ComplexType]]
	elementMap        *tplcontainer.SequenceMap[string, *Symbol[*proto.Element]]
	groupMap          *tplcontainer.SequenceMap[string, *Symbol[*proto.Group]]
	attributeMap      *tplcontainer.SequenceMap[string, *Symbol[*proto.Attribute]]
	attributeGroupMap *tplcontainer.SequenceMap[string, *Symbol[*proto.AttributeGroup]]
}

func NewSymbolMap() *SymbolMap {
	return &SymbolMap{
		simpleTypeMap:     tplcontainer.NewSequenceMap[string, *Symbol[*proto.SimpleType]](),
		complexTypeMap:    tplcontainer.NewSequenceMap[string, *Symbol[*proto.ComplexType]](),
		elementMap:        tplcontainer.NewSequenceMap[string, *Symbol[*proto.Element]](),
		groupMap:          tplcontainer.NewSequenceMap[string, *Symbol[*proto.Group]](),
		attributeMap:      tplcontainer.NewSequenceMap[string, *Symbol[*proto.Attribute]](),
		attributeGroupMap: tplcontainer.NewSequenceMap[string, *Symbol[*proto.AttributeGroup]](),
	}
}

func (sm *SymbolMap) addSimpleType(st *proto.SimpleType, fileName string) {
	if len(st.Name) == 0 {
		panic("empty simpleType name")
	}
	sm.simpleTypeMap.Set(st.Name, &Symbol[*proto.SimpleType]{
		Symbol:   st,
		FileName: fileName,
	})
}

func (sm *SymbolMap) addConplexType(ct *proto.ComplexType, fileName string) {
	if len(ct.Name) == 0 {
		panic("empty complexType name")
	}
	sm.complexTypeMap.Set(ct.Name, &Symbol[*proto.ComplexType]{
		Symbol:   ct,
		FileName: fileName,
	})
}

func (sm *SymbolMap) addElement(elem *proto.Element, fileName string) {
	if len(elem.Name) == 0 {
		panic("empty element name")
	}
	sm.elementMap.Set(elem.Name, &Symbol[*proto.Element]{
		Symbol:   elem,
		FileName: fileName,
	})
}

func (sm *SymbolMap) addGroup(grp *proto.Group, fileName string) {
	if len(grp.Name) == 0 {
		panic("empty group name")
	}
	sm.groupMap.Set(grp.Name, &Symbol[*proto.Group]{
		Symbol:   grp,
		FileName: fileName,
	})
}

func (sm *SymbolMap) addAttribute(attr *proto.Attribute, fileName string) {
	if len(attr.Name) == 0 {
		panic("empty attribute name")
	}
	sm.attributeMap.Set(attr.Name, &Symbol[*proto.Attribute]{
		Symbol:   attr,
		FileName: fileName,
	})
}

func (sm *SymbolMap) addAttributeGroup(attrGrp *proto.AttributeGroup, fileName string) {
	if len(attrGrp.Name) == 0 {
		panic("empty attributeGroup name")
	}
	sm.attributeGroupMap.Set(attrGrp.Name, &Symbol[*proto.AttributeGroup]{
		Symbol:   attrGrp,
		FileName: fileName,
	})
}

func (sm *SymbolMap) AddSymbolsFromSchema(schema *proto.Schema, fileName string) {
	for _, ct := range schema.ComplexTypeList {
		sm.addConplexType(ct, fileName)
	}

	for _, st := range schema.SimpleTypeList {
		sm.addSimpleType(st, fileName)
	}

	for _, elem := range schema.ElementList {
		sm.addElement(elem, fileName)
	}

	for _, grp := range schema.GroupList {
		sm.addGroup(grp, fileName)
	}

	for _, attr := range schema.AttributeList {
		sm.addAttribute(attr, fileName)
	}

	for _, attrGrp := range schema.AttributeGroupList {
		sm.addAttributeGroup(attrGrp, fileName)
	}
}

func addFromMap[SYMBOL any](src, dst *tplcontainer.SequenceMap[string, *Symbol[SYMBOL]]) {
	for _, name := range dst.Order() {
		if _, ok := dst.Get(name); ok {
			panic("duplicate symbol: " + name)
		}
		info := src.MustGet(name)
		dst.Set(name, info)
	}
}

func (sm *SymbolMap) AddSymbolsFromMap(symbs *SymbolMap) {
	addFromMap(sm.simpleTypeMap, symbs.simpleTypeMap)
	addFromMap(sm.complexTypeMap, symbs.complexTypeMap)
	addFromMap(sm.elementMap, symbs.elementMap)
	addFromMap(sm.groupMap, symbs.groupMap)
	addFromMap(sm.attributeMap, symbs.attributeMap)
	addFromMap(sm.attributeGroupMap, symbs.attributeGroupMap)
}

func mergeFromMap[SYMBOL any](dst, src *tplcontainer.SequenceMap[string, *Symbol[SYMBOL]], overwrite bool) {
	for _, name := range src.Order() {
		v := src.MustGet(name)
		if _, ok := dst.Get(name); !ok || overwrite {
			dst.Set(name, v)
		}
	}
}

// Merge 将符号表合并。允许重复的符号，可选择是否覆盖重复的符号
func (sm *SymbolMap) Merge(symbs *SymbolMap, overwrite bool) {
	mergeFromMap(sm.simpleTypeMap, symbs.simpleTypeMap, overwrite)
	mergeFromMap(sm.complexTypeMap, symbs.complexTypeMap, overwrite)
	mergeFromMap(sm.elementMap, symbs.elementMap, overwrite)
	mergeFromMap(sm.groupMap, symbs.groupMap, overwrite)
	mergeFromMap(sm.attributeMap, symbs.attributeMap, overwrite)
	mergeFromMap(sm.attributeGroupMap, symbs.attributeGroupMap, overwrite)
}

func (sm *SymbolMap) GetSimpleType(name string) (*Symbol[*proto.SimpleType], bool) {
	return sm.simpleTypeMap.Get(name)
}

func (sm *SymbolMap) GetComplexType(name string) (*Symbol[*proto.ComplexType], bool) {
	return sm.complexTypeMap.Get(name)
}

func (sm *SymbolMap) GetType(name string) (*Symbol[any], bool) {
	ct, ok := sm.GetComplexType(name)
	if ok {
		return &Symbol[any]{
			Symbol: ct.Symbol,
		}, true
	}

	st, ok := sm.GetSimpleType(name)
	if ok {
		return &Symbol[any]{
			Symbol: st.Symbol,
		}, true
	}
	return nil, false
}

func (sm *SymbolMap) GetElement(name string) (*Symbol[*proto.Element], bool) {
	return sm.elementMap.Get(name)
}

func (sm *SymbolMap) GetGroup(name string) (*Symbol[*proto.Group], bool) {
	return sm.groupMap.Get(name)
}

func (sm *SymbolMap) GetAttribute(name string) (*Symbol[*proto.Attribute], bool) {
	return sm.attributeMap.Get(name)
}

func (sm *SymbolMap) GetAttributeGroup(name string) (*Symbol[*proto.AttributeGroup], bool) {
	return sm.attributeGroupMap.Get(name)
}
