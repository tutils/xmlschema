package render

import (
	"io/fs"
	"net/url"
	"path/filepath"

	"github.com/tutils/xmlschema/proto"
	"github.com/tutils/xmlschema/tplcontainer"
)

/*

NamespaceMap {
	Namespace -> SymbolMap {
		SymbolName -> SymbolInfo {
			proto.Symbol
			FileName
		}
	}
}

FileMap {
    FileName -> FileScope {
        Namespace
        PrefixMap {
            Prefix -> Namespace
        }
        SymbolList
    }
}
*/

type loadState int

const (
	unloaded loadState = iota // 未加载
	loading                   // 加载中
	loaded                    // 已加载
)

type GlobalScope struct {
	namespaceMap *tplcontainer.SequenceMap[string, *SymbolMap]
	fileMap      *tplcontainer.SequenceMap[string, *FileScope]

	// [file, record]
	cache FileCache
}

func NewGlobalScope() *GlobalScope {
	gs := &GlobalScope{
		namespaceMap: tplcontainer.NewSequenceMap[string, *SymbolMap](),
		fileMap:      tplcontainer.NewSequenceMap[string, *FileScope](),

		cache: make(FileCache),
	}

	for _, fs := range []*FileScope{
		xmlFs,
		xsFs,
	} {
		gs.cache[fs.name] = fs.schema
		gs.addFileScope(fs)
	}

	return gs
}

func (gs *GlobalScope) addFileScope(fs *FileScope) {
	// 添加文件映射
	gs.fileMap.Set(fs.name, fs)

	// 合并符号表
	gs.mergeNamespaceMap(fs)
}

// mergeNamespaceMap 将文件域的符号表合并至全局符号表中
func (gs *GlobalScope) mergeNamespaceMap(fs *FileScope) {
	for _, ns := range fs.namespaceMap.Order() {
		if len(ns) == 0 {
			continue
		}

		fileSymbs := fs.namespaceMap.MustGet(ns)
		globalSymbs, ok := gs.namespaceMap.Get(ns)
		if !ok {
			globalSymbs = NewSymbolMap()
			gs.namespaceMap.Set(ns, globalSymbs)
		}
		globalSymbs.Merge(fileSymbs, false)
	}
}

func (gs *GlobalScope) LoadSchema(name string) *FileScope {
	u, err := url.Parse(name)
	if err != nil {
		panic(err)
	}
	if len(u.Scheme) == 0 {
		// 文件名标准化
		dir := filepath.Dir(name)
		name = filepath.Join(dir, filepath.Base(name))
	}

	if fs, ok := gs.fileMap.Get(name); ok {
		// 已经加载过该文件
		return fs
	}

	fs := newFileScope()
	fs.loadSchema(name, gs.cache)

	gs.addFileScope(fs)

	return fs
}

func (gs *GlobalScope) LoadSchemaFilesFromDirectory(dir string) {
	filepath.Walk(dir, func(path string, info fs.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		gs.LoadSchema(path)

		return nil
	})
}

func (gs *GlobalScope) GetSimpleType(namespace string, name string) (*Symbol[*proto.SimpleType], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetSimpleType(name)
}

func (gs *GlobalScope) GetComplexType(namespace string, name string) (*Symbol[*proto.ComplexType], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetComplexType(name)
}

func (gs *GlobalScope) GetType(namespace string, name string) (*Symbol[any], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetType(name)
}

func (gs *GlobalScope) GetElement(namespace string, name string) (*Symbol[*proto.Element], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetElement(name)
}

func (gs *GlobalScope) GetGroup(namespace string, name string) (*Symbol[*proto.Group], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetGroup(name)
}

func (gs *GlobalScope) GetAttribute(namespace string, name string) (*Symbol[*proto.Attribute], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetAttribute(name)
}

func (gs *GlobalScope) GetAttributeGroup(namespace string, name string) (*Symbol[*proto.AttributeGroup], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetAttributeGroup(name)
}

func (gs *GlobalScope) GetSimpleTypeInFile(nameWithPrefix string, fileName string) (string, *Symbol[*proto.SimpleType], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}
	prefix, symbol, ok := fs.GetSimpleType(nameWithPrefix)
	return prefix, symbol, ok
}

func (gs *GlobalScope) GetComplexTypeInFile(nameWithPrefix string, fileName string) (string, *Symbol[*proto.ComplexType], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}
	prefix, symbol, ok := fs.GetComplexType(nameWithPrefix)
	return prefix, symbol, ok
}

func (gs *GlobalScope) GetTypeInFile(nameWithPrefix string, fileName string) (string, *Symbol[any], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}
	prefix, symbol, ok := fs.GetType(nameWithPrefix)
	return prefix, symbol, ok
}

func (gs *GlobalScope) GetElementInFile(nameWithPrefix string, fileName string) (string, *Symbol[*proto.Element], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}
	prefix, symbol, ok := fs.GetElement(nameWithPrefix)
	return prefix, symbol, ok
}

func (gs *GlobalScope) GetGroupInFile(nameWithPrefix string, fileName string) (string, *Symbol[*proto.Group], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}
	prefix, symbol, ok := fs.GetGroup(nameWithPrefix)
	return prefix, symbol, ok
}

func (gs *GlobalScope) GetAttributeInFile(nameWithPrefix string, fileName string) (string, *Symbol[*proto.Attribute], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}
	prefix, symbol, ok := fs.GetAttribute(nameWithPrefix)
	return prefix, symbol, ok
}

func (gs *GlobalScope) GetAttributeGroupInFile(nameWithPrefix string, fileName string) (string, *Symbol[*proto.AttributeGroup], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}
	prefix, symbol, ok := fs.GetAttributeGroup(nameWithPrefix)
	return prefix, symbol, ok
}

func (gs *GlobalScope) GetElementByRef(elem *proto.Element, fileName string) (string, *Symbol[*proto.Element], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}

	prefix, symbol, ok := fs.GetElement(elem.Ref)
	if !ok {
		return prefix, nil, false
	}

	return prefix, symbol, true
}

func (gs *GlobalScope) GetGroupByRef(grp *proto.Group, fileName string) (string, *Symbol[*proto.Group], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}

	prefix, symbol, ok := fs.GetGroup(grp.Ref)
	if !ok {
		return prefix, nil, false
	}

	return prefix, symbol, true
}

func (gs *GlobalScope) GetAttributeByRef(attr *proto.Attribute, fileName string) (string, *Symbol[*proto.Attribute], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}

	prefix, symbol, ok := fs.GetAttribute(attr.Ref)
	if !ok {
		return prefix, nil, false
	}

	return prefix, symbol, true
}

func (gs *GlobalScope) GetAttributeGroupByRef(attrGrp *proto.AttributeGroup, fileName string) (string, *Symbol[*proto.AttributeGroup], bool) {
	fs, ok := gs.fileMap.Get(fileName)
	if !ok {
		return "", nil, false
	}

	prefix, symbol, ok := fs.GetAttributeGroup(attrGrp.Ref)
	if !ok {
		return prefix, nil, false
	}

	return prefix, symbol, true
}
