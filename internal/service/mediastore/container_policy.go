package mediastore

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/internal/client"
	"github.com/terraform-providers/terraform-provider-aws/internal/verify"
)

func ResourceContainerPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaStoreContainerPolicyPut,
		Read:   resourceAwsMediaStoreContainerPolicyRead,
		Update: resourceAwsMediaStoreContainerPolicyPut,
		Delete: resourceAwsMediaStoreContainerPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.ValidIAMPolicyJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
		},
	}
}

func resourceAwsMediaStoreContainerPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).MediaStoreConn

	input := &mediastore.PutContainerPolicyInput{
		ContainerName: aws.String(d.Get("container_name").(string)),
		Policy:        aws.String(d.Get("policy").(string)),
	}

	_, err := conn.PutContainerPolicy(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("container_name").(string))
	return resourceAwsMediaStoreContainerPolicyRead(d, meta)
}

func resourceAwsMediaStoreContainerPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).MediaStoreConn

	input := &mediastore.GetContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	resp, err := conn.GetContainerPolicy(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodeContainerNotFoundException, "") {
			log.Printf("[WARN] MediaContainer Policy %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodePolicyNotFoundException, "") {
			log.Printf("[WARN] MediaContainer Policy %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("container_name", d.Id())
	d.Set("policy", resp.Policy)
	return nil
}

func resourceAwsMediaStoreContainerPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).MediaStoreConn

	input := &mediastore.DeleteContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteContainerPolicy(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodeContainerNotFoundException, "") {
			return nil
		}
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodePolicyNotFoundException, "") {
			return nil
		}
		// if isAWSErr(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
		// 	return nil
		// }
		return err
	}

	return nil
}
