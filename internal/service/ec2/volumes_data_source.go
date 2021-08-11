package ec2

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/internal/client"
	"github.com/terraform-providers/terraform-provider-aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/internal/tags"
)

func DataSourceEBSVolumes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEbsVolumesRead,
		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),

			"tags": tags.TagsSchema(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsEbsVolumesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).EC2Conn

	req := &ec2.DescribeVolumesInput{}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		req.Filters = append(req.Filters, buildEC2CustomFilterList(
			filters.(*schema.Set),
		)...)
	}

	if len(req.Filters) == 0 {
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeVolumes %s\n", req)
	resp, err := conn.DescribeVolumes(req)
	if err != nil {
		return fmt.Errorf("error describing EC2 Volumes: %w", err)
	}

	if resp == nil || len(resp.Volumes) == 0 {
		return errors.New("no matching volumes found")
	}

	volumes := make([]string, 0)

	for _, volume := range resp.Volumes {
		volumes = append(volumes, *volume.VolumeId)
	}

	d.SetId(meta.(*client.AWSClient).Region)

	if err := d.Set("ids", volumes); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
