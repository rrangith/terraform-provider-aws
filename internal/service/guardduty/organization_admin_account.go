package guardduty

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/internal/client"
	"github.com/terraform-providers/terraform-provider-aws/internal/verify"
)

func ResourceOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyOrganizationAdminAccountCreate,
		Read:   resourceAwsGuardDutyOrganizationAdminAccountRead,
		Delete: resourceAwsGuardDutyOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"admin_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceAwsGuardDutyOrganizationAdminAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).GuardDutyConn

	adminAccountID := d.Get("admin_account_id").(string)

	input := &guardduty.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
	}

	_, err := conn.EnableOrganizationAdminAccount(input)

	if err != nil {
		return fmt.Errorf("error enabling GuardDuty Organization Admin Account (%s): %w", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	if _, err := waitAdminAccountEnabled(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for GuardDuty Organization Admin Account (%s) to enable: %w", d.Id(), err)
	}

	return resourceAwsGuardDutyOrganizationAdminAccountRead(d, meta)
}

func resourceAwsGuardDutyOrganizationAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).GuardDutyConn

	adminAccount, err := getGuardDutyOrganizationAdminAccount(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading GuardDuty Organization Admin Account (%s): %w", d.Id(), err)
	}

	if adminAccount == nil {
		log.Printf("[WARN] GuardDuty Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("admin_account_id", adminAccount.AdminAccountId)

	return nil
}

func resourceAwsGuardDutyOrganizationAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).GuardDutyConn

	input := &guardduty.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccount(input)

	if err != nil {
		return fmt.Errorf("error disabling GuardDuty Organization Admin Account (%s): %w", d.Id(), err)
	}

	if _, err := waitAdminAccountNotFound(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for GuardDuty Organization Admin Account (%s) to disable: %w", d.Id(), err)
	}

	return nil
}

func getGuardDutyOrganizationAdminAccount(conn *guardduty.GuardDuty, adminAccountID string) (*guardduty.AdminAccount, error) {
	input := &guardduty.ListOrganizationAdminAccountsInput{}
	var result *guardduty.AdminAccount

	err := conn.ListOrganizationAdminAccountsPages(input, func(page *guardduty.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, adminAccount := range page.AdminAccounts {
			if adminAccount == nil {
				continue
			}

			if aws.StringValue(adminAccount.AdminAccountId) == adminAccountID {
				result = adminAccount
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
