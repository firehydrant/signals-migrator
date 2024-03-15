package renderer

import (
	_ "embed"
	"fmt"

	"github.com/firehydrant/signals-migrator/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type datadog struct {
	apiKey       string
	appKey       string
	organization types.Organization
	file         *hclwrite.File
}

var (
	//go:embed datadog_payload.txt
	payload string
)

func NewDatadog(apiKey string, appKey string, org types.Organization) datadog {
	return datadog{
		apiKey:       apiKey,
		appKey:       appKey,
		organization: org,
		file:         hclwrite.NewEmptyFile(),
	}
}

func (r *datadog) Hcl() (string, error) {
	err := r.Build(r.file)
	if err != nil {
		return "", err
	}

	return string(r.file.Bytes()), nil
}

func (r *datadog) Build(f *hclwrite.File) error {
	rootBody := f.Body()

	_ = rootBody.AppendNewBlock("provider", []string{"datadog"}).Body()
	locals := rootBody.AppendNewBlock("locals", nil).Body()
	locals.SetAttributeValue("payload", cty.StringVal(payload))

	for _, t := range r.organization.Teams {
		r.webhook(rootBody, t)
	}

	return nil
}

func (r *datadog) webhook(f *hclwrite.Body, team *types.Team) {
	t := f.AppendNewBlock("resource", []string{"datadog_webhook", team.Slug}).Body()
	t.SetAttributeValue("name", cty.StringVal(fmt.Sprintf("firehydrant-%s", team.Slug)))
	t.SetAttributeTraversal("url", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "data",
		},
		hcl.TraverseAttr{
			Name: "firehydrant_team",
		},
		hcl.TraverseAttr{
			Name: team.Slug,
		},
		hcl.TraverseAttr{
			Name: "datadog_transpose_url",
		},
	})
	t.SetAttributeValue("encode_as", cty.StringVal("json"))
	t.SetAttributeTraversal("payload", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "local",
		},
		hcl.TraverseAttr{
			Name: "payload",
		},
	})
}
