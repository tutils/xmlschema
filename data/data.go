package data

import "embed"

//go:embed schemas/xml/*.xsd templates/*.tpl
var Content embed.FS
