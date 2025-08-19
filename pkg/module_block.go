package pkg

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ blockWithSchema = &ModuleBlock{}
var _ rootBlock = &ModuleBlock{}

type ModuleBlock struct {
	HclBlock     *HclBlock
	File         *hcl.File
	HeadMetaArgs Args
	TailMetaArgs Args
	RequiredArgs Args
	OptionalArgs Args
	dir          string
}

func BuildModuleBlock(block *HclBlock, dir string, file *HclFile) (*ModuleBlock, error) {
	b := &ModuleBlock{
		dir:      dir,
		HclBlock: block,
		File:     file.File,
	}
	err := buildArgs(b, block.Attributes())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *ModuleBlock) AutoFix() error {
	blockToFix := b.HclBlock
	singleLineBlock := blockToFix.isSingleLineBlock()
	empty := true
	attributes := blockToFix.WriteBlock.Body().Attributes()
	blockToFix.Clear()
	if b.HeadMetaArgs != nil {
		blockToFix.appendNewline()
		blockToFix.writeArgs(b.HeadMetaArgs.SortHeadMetaArgs(), attributes)
		empty = false
	}
	if b.RequiredArgs != nil || b.OptionalArgs != nil {
		blockToFix.appendNewline()
		blockToFix.writeArgs(b.RequiredArgs.SortByName(), attributes)
		blockToFix.writeArgs(b.OptionalArgs.SortByName(), attributes)
		empty = false
	}
	if b.TailMetaArgs != nil {
		blockToFix.appendNewline()
		empty = false
	}
	blockToFix.writeArgs(b.TailMetaArgs.SortByName(), attributes)

	if singleLineBlock && !empty {
		blockToFix.appendNewline()
	}
	return nil
}

func (b *ModuleBlock) addTailMetaArg(arg *Arg) {
	b.TailMetaArgs = append(b.TailMetaArgs, arg)
}

func (b *ModuleBlock) addTailMetaNestedBlock(nb *NestedBlock) {
}

func (b *ModuleBlock) file() *hcl.File {
	return b.File
}

func (b *ModuleBlock) path() []string {
	return []string{"module", b.HclBlock.Labels[0]}
}

func (b *ModuleBlock) schemaBlock() (*tfjson.SchemaBlock, error) {
	moduleName := b.HclBlock.Labels[0]
	modulePath := filepath.Join(b.dir, ".terraform", "modules", moduleName)
	exists, err := afero.Exists(Fs, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if module %s exists: %w", moduleName, err)
	}
	sourceAttr, diag := b.HclBlock.Attributes()["source"].Expr.Value(&hcl.EvalContext{})
	if diag.HasErrors() {
		return nil, fmt.Errorf("failed to eval `source` attribute: %w", diag)
	}
	if !exists {
		// A local module
		modulePath = sourceAttr.AsString()
		modulePath = filepath.Join(b.dir, modulePath)
	}

	module, diagnostics := tfconfig.LoadModule(modulePath)
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("failed to load module from source %s, local dir %s: %w", sourceAttr.AsString(), modulePath, diagnostics)
	}
	schemaBlock := &tfjson.SchemaBlock{
		Attributes:      make(map[string]*tfjson.SchemaAttribute),
		NestedBlocks:    make(map[string]*tfjson.SchemaBlockType),
		DescriptionKind: tfjson.SchemaDescriptionKindPlain,
	}
	for _, variable := range module.Variables {
		schemaBlock.Attributes[variable.Name] = &tfjson.SchemaAttribute{
			AttributeType: cty.DynamicPseudoType,
			Required:      variable.Required,
			Optional:      !variable.Required,
			Sensitive:     variable.Sensitive,
		}
	}
	return schemaBlock, nil
}

var moduleHeadMetaArgs = map[string]int{"for_each": 0, "count": 0, "source": 1, "version": 2, "providers": 3}
var moduleTailMetaArgs = map[string]int{"depends_on": 0}

func (b *ModuleBlock) isHeadMeta(argNameOrNestedBlockType string) bool {
	_, ok := moduleHeadMetaArgs[argNameOrNestedBlockType]
	return ok
}

func (b *ModuleBlock) isTailMeta(argNameOrNestedBlockType string) bool {
	_, ok := moduleTailMetaArgs[argNameOrNestedBlockType]
	return ok
}

func (b *ModuleBlock) addHeadMetaArg(arg *Arg) {
	b.HeadMetaArgs = append(b.HeadMetaArgs, arg)
}

func (b *ModuleBlock) addOptionalAttr(arg *Arg) {
	b.OptionalArgs = append(b.OptionalArgs, arg)
}

func (b *ModuleBlock) addRequiredAttr(arg *Arg) {
	b.RequiredArgs = append(b.RequiredArgs, arg)
}

func (b *ModuleBlock) addOptionalNestedBlock(nb *NestedBlock) {

}

func (b *ModuleBlock) addRequiredNestedBlock(nb *NestedBlock) {

}
