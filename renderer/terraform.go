package renderer

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type terraform struct {
	file          *hclwrite.File
	providerBlock *hclwrite.Body
}

func NewTerraform() (*terraform, error) {
	t := terraform{
		file: hclwrite.NewEmptyFile(),
	}
	if err := t.Build(t.file); err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *terraform) Hcl() (string, error) {
	return string(t.file.Bytes()), nil
}

func (t *terraform) Build(f *hclwrite.File) error {
	rootBody := f.Body()

	tfBody := rootBody.AppendNewBlock("terraform", nil).Body()
	t.providerBlock = tfBody.AppendNewBlock("required_providers", nil).Body()
	return nil
}

func (t *terraform) Provider(name string, source string, version string) {
	t.providerBlock.SetAttributeValue(name, cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal(source),
		"version": cty.StringVal(version),
	}))
}
