package pkg

import (
	"regexp"

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

var outputsFileRegex = regexp.MustCompile(`.*?outputs.*?\.tf$`)
var variablesFileRegex = regexp.MustCompile(`.*?variables.*?\.tf$`)

func (f *HclFile) AutoFix() error {
	if outputsFileRegex.MatchString(f.FileName) {
		outputsFile := BuildOutputsFile(f)
		if err := outputsFile.AutoFix(); err != nil {
			return err
		}
		return nil
	}
	if variablesFileRegex.MatchString(f.FileName) {
		variablesFile := BuildVariablesFile(f)
		return variablesFile.AutoFix()
	}
	for i, b := range f.Body.(*hclsyntax.Body).Blocks {
		hclBlock := f.GetBlock(i)
		var ab AutoFixBlock
		if _, ok := blockTypesWithSchema[b.Type]; ok {
			ab = BuildBlockWithSchema(hclBlock, f.File)
		}
		switch b.Type {
		case "module":
			{
				var err error
				ab, err = BuildModuleBlock(hclBlock, f.dir.path, f.File)
				if err != nil {
					return err
				}
			}
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
		case "variable":
			{
				f.dir.AppendBlockToFile("variables.tf", hclBlock)
				_ = f.RemoveBlock(hclBlock)
			}
		case "output":
			{
				f.dir.AppendBlockToFile("outputs.tf", hclBlock)
				_ = f.RemoveBlock(hclBlock)
			}
		}

		if ab == nil {
			continue
		}
		if err := ab.AutoFix(); err != nil {
			return err
		}
	}
	return nil
}

func (f *HclFile) appendNewline() {
	f.WriteFile.Body().AppendNewline()
}

func (f *HclFile) appendBlock(b *HclBlock) {
	f.WriteFile.Body().AppendBlock(b.WriteBlock)
}

func (f *HclFile) ClearWriteFile() {
	for name := range f.WriteFile.Body().Attributes() {
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

func (f *HclFile) RemoveBlock(b *HclBlock) bool {
	return f.WriteFile.Body().RemoveBlock(b.WriteBlock)
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
