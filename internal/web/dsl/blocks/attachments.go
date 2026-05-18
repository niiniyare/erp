package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// AttachmentsBlock renders a collapsible file attachment section.
// Files are uploaded via multipart to the document's attachments endpoint.
func AttachmentsBlock(sess ui.UISessionContext) ast.Node {
	return ast.SectionNode{
		Title:     "Attachments",
		Collapsed: true,
		Body: []ast.Node{
			ast.InputTextNode{
				Name:  "attachments",
				Label: "Files",
			},
		},
	}
}
