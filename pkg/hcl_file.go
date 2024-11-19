package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"strings"
)

type HclFile struct {
	*hcl.File
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
		if b.Type == "resource" || b.Type == "data" {
			resourceBlock := BuildResourceBlock(hclBlock, f.File)
			resourceBlock.AutoFix()
		} else if b.Type == "locals" {
			localsBlock := BuildLocalsBlock(hclBlock, f.File)
			localsBlock.AutoFix()
		} else if b.Type == "terraform" {
			terraformBlock := BuildTerraformBlock(hclBlock, f.File)
			terraformBlock.AutoFix()
		}
	}
}

func (f *HclFile) appendNewline() {
	f.WriteFile.Body().AppendNewline()
}

func (f *HclFile) appendBlock(b *HclBlock) {
	f.WriteFile.Body().AppendBlock(b.WriteBlock)
}
