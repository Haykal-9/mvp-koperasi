// Package main for embedding. This file lives at the module root so that
// //go:embed can reference the sibling "templates" and "static" directories.
package koperasifrontend

import "embed"

//go:embed templates static
var FS embed.FS
