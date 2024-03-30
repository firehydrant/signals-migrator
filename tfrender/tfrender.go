package tfrender

import (
	"fmt"
	"os"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type TFRender struct {
	f *hclwrite.File

	provider *hclwrite.Body
	root     *hclwrite.Body

	dir string
}

func New(dir string) (*TFRender, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("preparing output directory: %w", err)
	}

	f := hclwrite.NewEmptyFile()
	root := f.Body()
	provider := root.AppendNewBlock("terraform", nil).Body().AppendNewBlock("required_providers", nil).Body()

	return &TFRender{
		f:        f,
		provider: provider,
		root:     root,
		dir:      dir,
	}, nil
}

func (r *TFRender) DataFireHydrantUsers(users map[string]*pager.User) error {
	for _, u := range users {
		if u == nil {
			continue
		}

		block := r.root.AppendNewBlock("data", []string{"firehydrant_user", u.TFSlug()}).Body()
		block.SetAttributeValue("email", cty.StringVal(u.Email))
		r.root.AppendNewline()
	}
	return nil
}
