package tfrender_test

import (
	"testing"
)

func TestRenderOpsgenie(t *testing.T) {
	// Render Terraform configuration for slightly complex teams (with memberships) and schedules.
	t.Run("TeamWithSchedules", assertRenderPager)
}
