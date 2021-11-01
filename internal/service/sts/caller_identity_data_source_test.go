package sts_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSTSCallerIdentityDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, sts.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCallerIdentityConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckCallerIdentityAccountID("data.aws_sts_caller_identity.current"),
				),
			},
		},
	})
}

const testAccCheckAWSCallerIdentityConfig_basic = `
data "aws_sts_caller_identity" "current" {}
`