package pkg

import (
	"errors"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/matt-FFFFFF/tfpluginschema/tfplugin5"
	"github.com/matt-FFFFFF/tfpluginschema/tfplugin6"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// convertV6ResponseToProviderSchema converts a tfplugin6.GetProviderSchema_Response to tfjson.ProviderSchema
func convertV6ResponseToProviderSchema(resp *tfplugin6.GetProviderSchema_Response) (*tfjson.ProviderSchema, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	providerSchema := &tfjson.ProviderSchema{
		ResourceSchemas:          make(map[string]*tfjson.Schema),
		DataSourceSchemas:        make(map[string]*tfjson.Schema),
		EphemeralResourceSchemas: make(map[string]*tfjson.Schema),
		Functions:                make(map[string]*tfjson.FunctionSignature),
	}

	// Convert provider configuration schema
	if resp.Provider != nil {
		schema, err := convertV6SchemaToTFJSONSchema(resp.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed to convert provider schema: %w", err)
		}
		providerSchema.ConfigSchema = schema
	}

	// Convert resource schemas
	for name, schema := range resp.ResourceSchemas {
		convertedSchema, err := convertV6SchemaToTFJSONSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource schema %s: %w", name, err)
		}
		providerSchema.ResourceSchemas[name] = convertedSchema
	}

	// Convert data source schemas
	for name, schema := range resp.DataSourceSchemas {
		convertedSchema, err := convertV6SchemaToTFJSONSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert data source schema %s: %w", name, err)
		}
		providerSchema.DataSourceSchemas[name] = convertedSchema
	}

	// Convert ephemeral resource schemas
	for name, schema := range resp.EphemeralResourceSchemas {
		convertedSchema, err := convertV6SchemaToTFJSONSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ephemeral resource schema %s: %w", name, err)
		}
		providerSchema.EphemeralResourceSchemas[name] = convertedSchema
	}

	// Convert functions
	for name, function := range resp.Functions {
		convertedFunction, err := convertV6FunctionToTFJSONFunction(function)
		if err != nil {
			return nil, fmt.Errorf("failed to convert function %s: %w", name, err)
		}
		providerSchema.Functions[name] = convertedFunction
	}

	return providerSchema, nil
}

// convertV5ResponseToProviderSchema converts a tfplugin5.GetProviderSchema_Response to tfjson.ProviderSchema
func convertV5ResponseToProviderSchema(resp *tfplugin5.GetProviderSchema_Response) (*tfjson.ProviderSchema, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	providerSchema := &tfjson.ProviderSchema{
		ResourceSchemas:          make(map[string]*tfjson.Schema),
		DataSourceSchemas:        make(map[string]*tfjson.Schema),
		EphemeralResourceSchemas: make(map[string]*tfjson.Schema),
		Functions:                make(map[string]*tfjson.FunctionSignature),
	}

	// Convert provider configuration schema
	if resp.Provider != nil {
		schema, err := convertV5SchemaToTFJSONSchema(resp.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed to convert provider schema: %w", err)
		}
		providerSchema.ConfigSchema = schema
	}

	// Convert resource schemas
	for name, schema := range resp.ResourceSchemas {
		convertedSchema, err := convertV5SchemaToTFJSONSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource schema %s: %w", name, err)
		}
		providerSchema.ResourceSchemas[name] = convertedSchema
	}

	// Convert data source schemas
	for name, schema := range resp.DataSourceSchemas {
		convertedSchema, err := convertV5SchemaToTFJSONSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert data source schema %s: %w", name, err)
		}
		providerSchema.DataSourceSchemas[name] = convertedSchema
	}

	// Convert ephemeral resource schemas
	for name, schema := range resp.EphemeralResourceSchemas {
		convertedSchema, err := convertV5SchemaToTFJSONSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ephemeral resource schema %s: %w", name, err)
		}
		providerSchema.EphemeralResourceSchemas[name] = convertedSchema
	}

	// Convert functions
	for name, function := range resp.Functions {
		convertedFunction, err := convertV5FunctionToTFJSONFunction(function)
		if err != nil {
			return nil, fmt.Errorf("failed to convert function %s: %w", name, err)
		}
		providerSchema.Functions[name] = convertedFunction
	}

	return providerSchema, nil
}

// convertV6SchemaToTFJSONSchema converts a tfplugin6.Schema to tfjson.Schema
func convertV6SchemaToTFJSONSchema(schema *tfplugin6.Schema) (*tfjson.Schema, error) {
	if schema == nil {
		return nil, nil
	}

	version, err := safeInt64ToUint64(schema.GetVersion())
	if err != nil {
		return nil, fmt.Errorf("failed to convert schema version: %w", err)
	}

	tfjsonSchema := &tfjson.Schema{
		Version: version,
	}

	// Convert the block
	if schema.GetBlock() != nil {
		block, err := convertV6SchemaBlockToTFJSONSchemaBlock(schema.GetBlock())
		if err != nil {
			return nil, fmt.Errorf("failed to convert schema block: %w", err)
		}
		tfjsonSchema.Block = block
	}

	return tfjsonSchema, nil
}

// convertV5SchemaToTFJSONSchema converts a tfplugin5.Schema to tfjson.Schema
func convertV5SchemaToTFJSONSchema(schema *tfplugin5.Schema) (*tfjson.Schema, error) {
	if schema == nil {
		return nil, nil
	}

	version, err := safeInt64ToUint64(schema.GetVersion())
	if err != nil {
		return nil, fmt.Errorf("failed to convert schema version: %w", err)
	}

	tfjsonSchema := &tfjson.Schema{
		Version: version,
	}

	// Convert the block
	if schema.GetBlock() != nil {
		block, err := convertV5SchemaBlockToTFJSONSchemaBlock(schema.GetBlock())
		if err != nil {
			return nil, fmt.Errorf("failed to convert schema block: %w", err)
		}
		tfjsonSchema.Block = block
	}

	return tfjsonSchema, nil
}

// convertV6FunctionToTFJSONFunction converts a tfplugin6.Function to tfjson.FunctionSignature
func convertV6FunctionToTFJSONFunction(function *tfplugin6.Function) (*tfjson.FunctionSignature, error) {
	if function == nil {
		return nil, nil
	}

	signature := &tfjson.FunctionSignature{
		Description:        function.GetDescription(),
		Summary:            function.GetSummary(),
		DeprecationMessage: function.GetDeprecationMessage(),
		Parameters:         make([]*tfjson.FunctionParameter, 0, len(function.GetParameters())),
	}

	// Convert parameters
	for _, param := range function.GetParameters() {
		convertedParam, err := convertV6FunctionParameterToTFJSON(param)
		if err != nil {
			return nil, fmt.Errorf("failed to convert function parameter: %w", err)
		}
		signature.Parameters = append(signature.Parameters, convertedParam)
	}

	// Convert variadic parameter
	if function.GetVariadicParameter() != nil {
		convertedParam, err := convertV6FunctionParameterToTFJSON(function.GetVariadicParameter())
		if err != nil {
			return nil, fmt.Errorf("failed to convert variadic parameter: %w", err)
		}
		signature.VariadicParameter = convertedParam
	}

	// Convert return type
	if function.GetReturn() != nil && function.GetReturn().GetType() != nil {
		returnType, err := ctyjson.UnmarshalType(function.GetReturn().GetType())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal return type: %w", err)
		}
		signature.ReturnType = returnType
	} else {
		signature.ReturnType = cty.DynamicPseudoType
	}

	return signature, nil
}

// convertV5FunctionToTFJSONFunction converts a tfplugin5.Function to tfjson.FunctionSignature
func convertV5FunctionToTFJSONFunction(function *tfplugin5.Function) (*tfjson.FunctionSignature, error) {
	if function == nil {
		return nil, nil
	}

	signature := &tfjson.FunctionSignature{
		Description:        function.GetDescription(),
		Summary:            function.GetSummary(),
		DeprecationMessage: function.GetDeprecationMessage(),
		Parameters:         make([]*tfjson.FunctionParameter, 0, len(function.GetParameters())),
	}

	// Convert parameters
	for _, param := range function.GetParameters() {
		convertedParam, err := convertV5FunctionParameterToTFJSON(param)
		if err != nil {
			return nil, fmt.Errorf("failed to convert function parameter: %w", err)
		}
		signature.Parameters = append(signature.Parameters, convertedParam)
	}

	// Convert variadic parameter
	if function.GetVariadicParameter() != nil {
		convertedParam, err := convertV5FunctionParameterToTFJSON(function.GetVariadicParameter())
		if err != nil {
			return nil, fmt.Errorf("failed to convert variadic parameter: %w", err)
		}
		signature.VariadicParameter = convertedParam
	}

	// Convert return type
	if function.GetReturn() != nil && function.GetReturn().GetType() != nil {
		returnType, err := ctyjson.UnmarshalType(function.GetReturn().GetType())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal return type: %w", err)
		}
		signature.ReturnType = returnType
	} else {
		signature.ReturnType = cty.DynamicPseudoType
	}

	return signature, nil
}

// convertV6FunctionParameterToTFJSON converts a tfplugin6.Function_Parameter to tfjson.FunctionParameter
func convertV6FunctionParameterToTFJSON(param *tfplugin6.Function_Parameter) (*tfjson.FunctionParameter, error) {
	if param == nil {
		return nil, nil
	}

	parameter := &tfjson.FunctionParameter{
		Name:        param.GetName(),
		Description: param.GetDescription(),
		IsNullable:  param.GetAllowNullValue(),
	}

	// Convert parameter type
	if param.GetType() != nil {
		paramType, err := ctyjson.UnmarshalType(param.GetType())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameter type: %w", err)
		}
		parameter.Type = paramType
	} else {
		parameter.Type = cty.DynamicPseudoType
	}

	return parameter, nil
}

// convertV5FunctionParameterToTFJSON converts a tfplugin5.Function_Parameter to tfjson.FunctionParameter
func convertV5FunctionParameterToTFJSON(param *tfplugin5.Function_Parameter) (*tfjson.FunctionParameter, error) {
	if param == nil {
		return nil, nil
	}

	parameter := &tfjson.FunctionParameter{
		Name:        param.GetName(),
		Description: param.GetDescription(),
		IsNullable:  param.GetAllowNullValue(),
	}

	// Convert parameter type
	if param.GetType() != nil {
		paramType, err := ctyjson.UnmarshalType(param.GetType())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameter type: %w", err)
		}
		parameter.Type = paramType
	} else {
		parameter.Type = cty.DynamicPseudoType
	}

	return parameter, nil
}

// convertV6SchemaBlockToTFJSONSchemaBlock converts a tfplugin6.Schema_Block to tfjson.SchemaBlock
func convertV6SchemaBlockToTFJSONSchemaBlock(block *tfplugin6.Schema_Block) (*tfjson.SchemaBlock, error) {
	if block == nil {
		return nil, nil
	}

	schemaBlock := &tfjson.SchemaBlock{
		Description:     block.GetDescription(),
		DescriptionKind: convertV6StringKindToTFJSONDescriptionKind(block.GetDescriptionKind()),
		Deprecated:      block.GetDeprecated(),
		Attributes:      make(map[string]*tfjson.SchemaAttribute),
		NestedBlocks:    make(map[string]*tfjson.SchemaBlockType),
	}

	// Convert attributes
	for _, attr := range block.GetAttributes() {
		convertedAttr, err := convertV6SchemaAttributeToTFJSONSchemaAttribute(attr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attribute %s: %w", attr.GetName(), err)
		}
		schemaBlock.Attributes[attr.GetName()] = convertedAttr
	}

	// Convert nested blocks
	for _, nestedBlock := range block.GetBlockTypes() {
		convertedBlock, err := convertV6SchemaNestedBlockToTFJSONSchemaBlockType(nestedBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to convert nested block %s: %w", nestedBlock.GetTypeName(), err)
		}
		schemaBlock.NestedBlocks[nestedBlock.GetTypeName()] = convertedBlock
	}

	return schemaBlock, nil
}

// convertV5SchemaBlockToTFJSONSchemaBlock converts a tfplugin5.Schema_Block to tfjson.SchemaBlock
func convertV5SchemaBlockToTFJSONSchemaBlock(block *tfplugin5.Schema_Block) (*tfjson.SchemaBlock, error) {
	if block == nil {
		return nil, nil
	}

	schemaBlock := &tfjson.SchemaBlock{
		Description:     block.GetDescription(),
		DescriptionKind: convertV5StringKindToTFJSONDescriptionKind(block.GetDescriptionKind()),
		Deprecated:      block.GetDeprecated(),
		Attributes:      make(map[string]*tfjson.SchemaAttribute),
		NestedBlocks:    make(map[string]*tfjson.SchemaBlockType),
	}

	// Convert attributes
	for _, attr := range block.GetAttributes() {
		convertedAttr, err := convertV5SchemaAttributeToTFJSONSchemaAttribute(attr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attribute %s: %w", attr.GetName(), err)
		}
		schemaBlock.Attributes[attr.GetName()] = convertedAttr
	}

	// Convert nested blocks
	for _, nestedBlock := range block.GetBlockTypes() {
		convertedBlock, err := convertV5SchemaNestedBlockToTFJSONSchemaBlockType(nestedBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to convert nested block %s: %w", nestedBlock.GetTypeName(), err)
		}
		schemaBlock.NestedBlocks[nestedBlock.GetTypeName()] = convertedBlock
	}

	return schemaBlock, nil
}

// Helper conversion functions for string kinds
func convertV6StringKindToTFJSONDescriptionKind(kind tfplugin6.StringKind) tfjson.SchemaDescriptionKind {
	switch kind {
	case tfplugin6.StringKind_MARKDOWN:
		return tfjson.SchemaDescriptionKindMarkdown
	default:
		return tfjson.SchemaDescriptionKindPlain
	}
}

func convertV5StringKindToTFJSONDescriptionKind(kind tfplugin5.StringKind) tfjson.SchemaDescriptionKind {
	switch kind {
	case tfplugin5.StringKind_MARKDOWN:
		return tfjson.SchemaDescriptionKindMarkdown
	default:
		return tfjson.SchemaDescriptionKindPlain
	}
}

// convertV6SchemaAttributeToTFJSONSchemaAttribute converts a tfplugin6.Schema_Attribute to tfjson.SchemaAttribute
func convertV6SchemaAttributeToTFJSONSchemaAttribute(attr *tfplugin6.Schema_Attribute) (*tfjson.SchemaAttribute, error) {
	if attr == nil {
		return nil, nil
	}

	schemaAttr := &tfjson.SchemaAttribute{
		Description:     attr.GetDescription(),
		DescriptionKind: convertV6StringKindToTFJSONDescriptionKind(attr.GetDescriptionKind()),
		Deprecated:      attr.GetDeprecated(),
		Required:        attr.GetRequired(),
		Optional:        attr.GetOptional(),
		Computed:        attr.GetComputed(),
		Sensitive:       attr.GetSensitive(),
	}

	// Convert attribute type if present
	if attr.GetType() != nil {
		attrType, err := ctyjson.UnmarshalType(attr.GetType())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal attribute type: %w", err)
		}
		schemaAttr.AttributeType = attrType
	}

	// Convert nested type if present
	if attr.GetNestedType() != nil {
		nestedType, err := convertV6SchemaNestedAttributeTypeToTFJSONSchemaNestedAttributeType(attr.GetNestedType())
		if err != nil {
			return nil, fmt.Errorf("failed to convert nested attribute type: %w", err)
		}
		schemaAttr.AttributeNestedType = nestedType
	}

	return schemaAttr, nil
}

// convertV5SchemaAttributeToTFJSONSchemaAttribute converts a tfplugin5.Schema_Attribute to tfjson.SchemaAttribute
func convertV5SchemaAttributeToTFJSONSchemaAttribute(attr *tfplugin5.Schema_Attribute) (*tfjson.SchemaAttribute, error) {
	if attr == nil {
		return nil, nil
	}

	schemaAttr := &tfjson.SchemaAttribute{
		Description:     attr.GetDescription(),
		DescriptionKind: convertV5StringKindToTFJSONDescriptionKind(attr.GetDescriptionKind()),
		Deprecated:      attr.GetDeprecated(),
		Required:        attr.GetRequired(),
		Optional:        attr.GetOptional(),
		Computed:        attr.GetComputed(),
		Sensitive:       attr.GetSensitive(),
		WriteOnly:       attr.GetWriteOnly(),
	}

	// Convert attribute type if present
	if attr.GetType() != nil {
		attrType, err := ctyjson.UnmarshalType(attr.GetType())
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal attribute type: %w", err)
		}
		schemaAttr.AttributeType = attrType
	}

	// tfplugin5 doesn't have nested types, so AttributeNestedType remains nil

	return schemaAttr, nil
}

// convertV6SchemaNestedBlockToTFJSONSchemaBlockType converts a tfplugin6.Schema_NestedBlock to tfjson.SchemaBlockType
func convertV6SchemaNestedBlockToTFJSONSchemaBlockType(nestedBlock *tfplugin6.Schema_NestedBlock) (*tfjson.SchemaBlockType, error) {
	if nestedBlock == nil {
		return nil, nil
	}
	minItems, err := safeInt64ToUint64(nestedBlock.GetMinItems())
	if err != nil {
		return nil, fmt.Errorf("failed to convert min items: %w", err)
	}
	maxItems, err := safeInt64ToUint64(nestedBlock.GetMaxItems())
	if err != nil {
		return nil, fmt.Errorf("failed to convert max items: %w", err)
	}

	blockType := &tfjson.SchemaBlockType{
		NestingMode: convertV6NestingModeToTFJSONNestingMode(nestedBlock.GetNesting()),
		MinItems:    minItems,
		MaxItems:    maxItems,
	}

	// Convert the nested block
	if nestedBlock.GetBlock() != nil {
		block, err := convertV6SchemaBlockToTFJSONSchemaBlock(nestedBlock.GetBlock())
		if err != nil {
			return nil, fmt.Errorf("failed to convert nested block: %w", err)
		}
		blockType.Block = block
	}

	return blockType, nil
}

// convertV5SchemaNestedBlockToTFJSONSchemaBlockType converts a tfplugin5.Schema_NestedBlock to tfjson.SchemaBlockType
func convertV5SchemaNestedBlockToTFJSONSchemaBlockType(nestedBlock *tfplugin5.Schema_NestedBlock) (*tfjson.SchemaBlockType, error) {
	if nestedBlock == nil {
		return nil, nil
	}

	minItems, err := safeInt64ToUint64(nestedBlock.GetMinItems())
	if err != nil {
		return nil, fmt.Errorf("failed to convert min items: %w", err)
	}
	maxItems, err := safeInt64ToUint64(nestedBlock.GetMaxItems())
	if err != nil {
		return nil, fmt.Errorf("failed to convert max items: %w", err)
	}

	blockType := &tfjson.SchemaBlockType{
		NestingMode: convertV5NestingModeToTFJSONNestingMode(nestedBlock.GetNesting()),
		MinItems:    minItems,
		MaxItems:    maxItems,
	}

	// Convert the nested block
	if nestedBlock.GetBlock() != nil {
		block, err := convertV5SchemaBlockToTFJSONSchemaBlock(nestedBlock.GetBlock())
		if err != nil {
			return nil, fmt.Errorf("failed to convert nested block: %w", err)
		}
		blockType.Block = block
	}

	return blockType, nil
}

// Helper conversion functions for nesting modes
func convertV6NestingModeToTFJSONNestingMode(nesting tfplugin6.Schema_NestedBlock_NestingMode) tfjson.SchemaNestingMode {
	switch nesting {
	case tfplugin6.Schema_NestedBlock_SINGLE:
		return tfjson.SchemaNestingModeSingle
	case tfplugin6.Schema_NestedBlock_LIST:
		return tfjson.SchemaNestingModeList
	case tfplugin6.Schema_NestedBlock_SET:
		return tfjson.SchemaNestingModeSet
	case tfplugin6.Schema_NestedBlock_MAP:
		return tfjson.SchemaNestingModeMap
	default:
		return tfjson.SchemaNestingModeSingle // default to single
	}
}

func convertV5NestingModeToTFJSONNestingMode(nesting tfplugin5.Schema_NestedBlock_NestingMode) tfjson.SchemaNestingMode {
	switch nesting {
	case tfplugin5.Schema_NestedBlock_SINGLE:
		return tfjson.SchemaNestingModeSingle
	case tfplugin5.Schema_NestedBlock_LIST:
		return tfjson.SchemaNestingModeList
	case tfplugin5.Schema_NestedBlock_SET:
		return tfjson.SchemaNestingModeSet
	case tfplugin5.Schema_NestedBlock_MAP:
		return tfjson.SchemaNestingModeMap
	default:
		return tfjson.SchemaNestingModeSingle // default to single
	}
}

// convertV6SchemaNestedAttributeTypeToTFJSONSchemaNestedAttributeType converts a tfplugin6.Schema_Object to tfjson.SchemaNestedAttributeType
func convertV6SchemaNestedAttributeTypeToTFJSONSchemaNestedAttributeType(obj *tfplugin6.Schema_Object) (*tfjson.SchemaNestedAttributeType, error) {
	if obj == nil {
		return nil, nil
	}

	nestedType := &tfjson.SchemaNestedAttributeType{
		NestingMode: convertV6ObjectNestingModeToTFJSONNestingMode(obj.GetNesting()),
		Attributes:  make(map[string]*tfjson.SchemaAttribute),
	}

	// Convert attributes
	for _, attr := range obj.GetAttributes() {
		convertedAttr, err := convertV6SchemaAttributeToTFJSONSchemaAttribute(attr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert nested attribute %s: %w", attr.GetName(), err)
		}
		nestedType.Attributes[attr.GetName()] = convertedAttr
	}
	var err error

	// Set min/max items if applicable
	if obj.GetMinItems() > 0 {
		if nestedType.MinItems, err = safeInt64ToUint64(obj.GetMinItems()); err != nil {
			return nil, fmt.Errorf("failed to convert min items: %w", err)
		}
	}
	if obj.GetMaxItems() > 0 {
		if nestedType.MaxItems, err = safeInt64ToUint64(obj.GetMaxItems()); err != nil {
			return nil, fmt.Errorf("failed to convert max items: %w", err)
		}
	}

	return nestedType, nil
}

// Helper conversion function for object nesting mode
func convertV6ObjectNestingModeToTFJSONNestingMode(nesting tfplugin6.Schema_Object_NestingMode) tfjson.SchemaNestingMode {
	switch nesting {
	case tfplugin6.Schema_Object_SINGLE:
		return tfjson.SchemaNestingModeSingle
	case tfplugin6.Schema_Object_LIST:
		return tfjson.SchemaNestingModeList
	case tfplugin6.Schema_Object_SET:
		return tfjson.SchemaNestingModeSet
	case tfplugin6.Schema_Object_MAP:
		return tfjson.SchemaNestingModeMap
	default:
		return tfjson.SchemaNestingModeSingle // default to single
	}
}

// SafeInt64ToUint64 converts an int64 to a uint64, returning an error if the input is negative.
func safeInt64ToUint64(val int64) (uint64, error) {
	// 1. Check if the value is negative.
	if val < 0 {
		return 0, errors.New("cannot convert a negative int64 to uint64")
	}
	// 2. If non-negative, the cast is safe.
	return uint64(val), nil
}
