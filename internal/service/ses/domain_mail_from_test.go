package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/internal/acctest"
	"github.com/terraform-providers/terraform-provider-aws/internal/client"
)

func TestAccAWSSESDomainMailFrom_basic(t *testing.T) {
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain1 := dn.Subdomain("bounce1").String()
	mailFromDomain2 := dn.Subdomain("bounce2").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESDomainMailFromConfig(domain, mailFromDomain1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1),
				),
			},
			{
				Config: testAccAwsSESDomainMailFromConfig(domain, mailFromDomain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSESDomainMailFrom_disappears(t *testing.T) {
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESDomainMailFromConfig(domain, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainMailFromExists(resourceName),
					testAccCheckAwsSESDomainMailFromDisappears(domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSESDomainMailFrom_disappears_Identity(t *testing.T) {
	dn := acctest.RandomDomain()
	domain := dn.String()
	mailFromDomain := dn.Subdomain("bounce").String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESDomainMailFromConfig(domain, mailFromDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainMailFromExists(resourceName),
					testAccCheckAwsSESDomainIdentityDisappears(domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSESDomainMailFrom_behaviorOnMxFailure(t *testing.T) {
	domain := acctest.RandomDomain().String()
	resourceName := "aws_ses_domain_mail_from.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESDomainMailFromConfig_behaviorOnMxFailure(domain, ses.BehaviorOnMXFailureUseDefaultValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
				),
			},
			{
				Config: testAccAwsSESDomainMailFromConfig_behaviorOnMxFailure(domain, ses.BehaviorOnMXFailureRejectMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainMailFromExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureRejectMessage),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsSESDomainMailFromExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := acctest.Provider.Meta().(*client.AWSClient).SESConn

		params := &ses.GetIdentityMailFromDomainAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityMailFromDomainAttributes(params)
		if err != nil {
			return err
		}

		if response.MailFromDomainAttributes[domain] == nil {
			return fmt.Errorf("SES Domain MAIL FROM %s not found in AWS", domain)
		}

		return nil
	}
}

func testAccCheckSESDomainMailFromDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*client.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_domain_mail_from" {
			continue
		}

		input := &ses.GetIdentityMailFromDomainAttributesInput{
			Identities: []*string{aws.String(rs.Primary.ID)},
		}

		out, err := conn.GetIdentityMailFromDomainAttributes(input)
		if err != nil {
			return fmt.Errorf("error fetching MAIL FROM domain attributes: %s", err)
		}
		if v, ok := out.MailFromDomainAttributes[rs.Primary.ID]; ok && v.MailFromDomain != nil && *v.MailFromDomain != "" {
			return fmt.Errorf("MAIL FROM domain was not removed, found: %s", *v.MailFromDomain)
		}
	}

	return nil
}

func testAccCheckAwsSESDomainMailFromDisappears(identity string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*client.AWSClient).SESConn

		input := &ses.SetIdentityMailFromDomainInput{
			Identity:       aws.String(identity),
			MailFromDomain: nil,
		}

		_, err := conn.SetIdentityMailFromDomain(input)

		return err
	}
}

func testAccAwsSESDomainMailFromConfig(domain, mailFromDomain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_domain_mail_from" "test" {
  domain           = aws_ses_domain_identity.test.domain
  mail_from_domain = "%s"
}
`, domain, mailFromDomain)
}

func testAccAwsSESDomainMailFromConfig_behaviorOnMxFailure(domain, behaviorOnMxFailure string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_domain_mail_from" "test" {
  behavior_on_mx_failure = "%s"
  domain                 = aws_ses_domain_identity.test.domain
  mail_from_domain       = "bounce.${aws_ses_domain_identity.test.domain}"
}
`, domain, behaviorOnMxFailure)
}
