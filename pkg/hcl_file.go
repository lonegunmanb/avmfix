package pkg

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type HclFile struct {
	*hcl.File
	dir       *directory
	WriteFile *hclwrite.File
	FileName  string
}

func ParseConfig(config []byte, filename string) (*HclFile, hcl.Diagnostics) {
	file, rDiag := hclsyntax.ParseConfig(config, filename, hcl.InitialPos)
	writeFile, wDiag := hclwrite.ParseConfig(config, filename, hcl.InitialPos)
	if rDiag.HasErrors() || wDiag.HasErrors() {
		return nil, rDiag.Extend(wDiag)
	}
	return &HclFile{
		File:      file,
		WriteFile: writeFile,
		FileName:  filename,
	}, hcl.Diagnostics{}
}

func (f *HclFile) GetBlock(i int) *HclBlock {
	block := f.Body.(*hclsyntax.Body).Blocks[i]
	writeBlock := f.WriteFile.Body().Blocks()[i]
	return NewHclBlock(block, writeBlock)
}

func (f *HclFile) AutoFix() {
	if strings.HasSuffix(f.FileName, "outputs.tf") {
		outputsFile := BuildOutputsFile(f)
		outputsFile.AutoFix()
		return
	}
	if strings.HasSuffix(f.FileName, "variables.tf") {
		variablesFile := BuildVariablesFile(f)
		variablesFile.AutoFix()
		return
	}
	for i, b := range f.Body.(*hclsyntax.Body).Blocks {
		hclBlock := f.GetBlock(i)
		var ab AutoFixBlock
		if _, ok := blockTypesWithSchema[b.Type]; ok {
			ab = BuildBlockWithSchema(hclBlock, f.File)
		}
		switch b.Type {
		case "moved":
			{
				ab = BuildMovedBlock(hclBlock, f.File)
			}
		case "removed":
			{
				ab = BuildRemovedBlock(hclBlock, f.File)
			}
		case "locals":
			{
				ab = BuildLocalsBlock(hclBlock, f.File)
			}
		case "terraform":
			{
				ab = BuildTerraformBlock(hclBlock, f.File)
			}
		}

		if ab != nil {
			ab.AutoFix()
		}
	}
}

func (f *HclFile) appendNewline() {
	f.WriteFile.Body().AppendNewline()
}

func (f *HclFile) appendBlock(b *HclBlock) {
	f.WriteFile.Body().AppendBlock(b.WriteBlock)
}

func (f *HclFile) ClearWriteFile() {
	for name, _ := range f.WriteFile.Body().Attributes() {
		f.WriteFile.Body().RemoveAttribute(name)
	}
	for _, b := range f.WriteFile.Body().Blocks() {
		f.WriteFile.Body().RemoveBlock(b)
	}
	// There might be some seperated comments, like tflint ignore annotation in the head, we must preserve them.
	tokens := f.WriteFile.BuildTokens(nil)
	newTokens := f.trimRedundantNewLines(tokens)

	f.WriteFile.Body().Clear()
	f.WriteFile.Body().AppendUnstructuredTokens(newTokens)
}

func (f *HclFile) trimRedundantNewLines(tokens hclwrite.Tokens) hclwrite.Tokens {
	var newTokens hclwrite.Tokens
	for i := 0; i < len(tokens)-1; i++ {
		current := tokens[i]
		next := tokens[i+1]
		if (current.Type == hclsyntax.TokenNewline) && (next.Type == hclsyntax.TokenQuotedNewline || next.Type == hclsyntax.TokenNewline) {
			continue
		}
		newTokens = append(newTokens, current)
	}
	tokens = newTokens
	firstNonNewLine := len(tokens)
	for i := 0; i < len(tokens)-1; i++ {
		if tokens[i].Type != hclsyntax.TokenNewline {
			firstNonNewLine = i
			break
		}
	}
	newTokens = tokens[firstNonNewLine:]
	return newTokens
}

func (f *HclFile) endWithNewLine() bool {
	tokens := f.WriteFile.BuildTokens(nil)
	if len(tokens) == 0 || tokens[0].Type == hclsyntax.TokenEOF {
		return true
	}
	return tokens[len(tokens)-2].Type == hclsyntax.TokenNewline
}
