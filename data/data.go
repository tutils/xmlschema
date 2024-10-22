package data

import "embed"

//go:embed schemas/xml/*.xsd schemas/dc/*.xsd templates/*.tpl
var Content embed.FS

//go:embed example/e_person.xml
var EPerson []byte

//go:embed example/slide1.xml
var PSld []byte
