package qdiwrap

import _ "embed"

//go:embed templates/header.tmpl
var headerTemplate string

//go:embed templates/body.tmpl
var bodyTemplate string
