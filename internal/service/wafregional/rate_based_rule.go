package wafregional

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/internal/client"
	"github.com/terraform-providers/terraform-provider-aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/internal/tags"
)

func ResourceRateBasedRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalRateBasedRuleCreate,
		Read:   resourceAwsWafRegionalRateBasedRuleRead,
		Update: resourceAwsWafRegionalRateBasedRuleUpdate,
		Delete: resourceAwsWafRegionalRateBasedRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metric_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validMetricName,
			},
			"predicate": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"negated": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"data_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(wafregional.PredicateType_Values(), false),
						},
					},
				},
			},
			"rate_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rate_limit": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(100),
			},
			"tags":     tags.TagsSchema(),
			"tags_all": tags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: tags.SetTagsDiff,
	}
}

func resourceAwsWafRegionalRateBasedRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).WAFRegionalConn
	defaultTagsConfig := meta.(*client.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))
	region := meta.(*client.AWSClient).Region

	wr := newWafRegionalRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateRateBasedRuleInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(d.Get("name").(string)),
			RateKey:     aws.String(d.Get("rate_key").(string)),
			RateLimit:   aws.Int64(int64(d.Get("rate_limit").(int))),
		}

		if len(tags) > 0 {
			params.Tags = tags.IgnoreAws().WafregionalTags()
		}

		return conn.CreateRateBasedRule(params)
	})
	if err != nil {
		return fmt.Errorf("Error creating WAF Regional Rate Based Rule (%s): %s", d.Id(), err)
	}
	resp := out.(*waf.CreateRateBasedRuleOutput)
	d.SetId(aws.StringValue(resp.Rule.RuleId))

	newPredicates := d.Get("predicate").(*schema.Set).List()
	if len(newPredicates) > 0 {
		noPredicates := []interface{}{}
		err := updateWafRateBasedRuleResourceWR(d.Id(), noPredicates, newPredicates, d.Get("rate_limit"), conn, region)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rate Based Rule: %s", err)
		}
	}

	return resourceAwsWafRegionalRateBasedRuleRead(d, meta)
}

func resourceAwsWafRegionalRateBasedRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).WAFRegionalConn
	defaultTagsConfig := meta.(*client.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*client.AWSClient).IgnoreTagsConfig

	params := &waf.GetRateBasedRuleInput{
		RuleId: aws.String(d.Id()),
	}

	resp, err := conn.GetRateBasedRule(params)
	if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
		log.Printf("[WARN] WAF Regional Rate Based Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting WAF Regional Rate Based Rule (%s): %s", d.Id(), err)
	}

	var predicates []map[string]interface{}

	for _, predicateSet := range resp.Rule.MatchPredicates {
		predicates = append(predicates, map[string]interface{}{
			"negated": *predicateSet.Negated,
			"type":    *predicateSet.Type,
			"data_id": *predicateSet.DataId,
		})
	}

	arn := arn.ARN{
		AccountID: meta.(*client.AWSClient).AccountID,
		Partition: meta.(*client.AWSClient).Partition,
		Region:    meta.(*client.AWSClient).Region,
		Resource:  fmt.Sprintf("ratebasedrule/%s", d.Id()),
		Service:   "waf-regional",
	}.String()
	d.Set("arn", arn)

	tagList, err := keyvaluetags.WafregionalListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("Failed to get WAF Regional Rated Based Rule parameter tags for %s: %s", d.Get("name"), err)
	}

	tags := tagList.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("predicate", predicates)
	d.Set("name", resp.Rule.Name)
	d.Set("metric_name", resp.Rule.MetricName)
	d.Set("rate_key", resp.Rule.RateKey)
	d.Set("rate_limit", resp.Rule.RateLimit)

	return nil
}

func resourceAwsWafRegionalRateBasedRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).WAFRegionalConn
	region := meta.(*client.AWSClient).Region

	if d.HasChanges("predicate", "rate_limit") {
		o, n := d.GetChange("predicate")
		oldP, newP := o.(*schema.Set).List(), n.(*schema.Set).List()
		rateLimit := d.Get("rate_limit")

		err := updateWafRateBasedRuleResourceWR(d.Id(), oldP, newP, rateLimit, conn, region)
		if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAF Regional Rate Based Rule (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error updating WAF Regional Rate Based Rule Predicates (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.WafregionalUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWafRegionalRateBasedRuleRead(d, meta)
}

func resourceAwsWafRegionalRateBasedRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*client.AWSClient).WAFRegionalConn
	region := meta.(*client.AWSClient).Region

	oldPredicates := d.Get("predicate").(*schema.Set).List()
	if len(oldPredicates) > 0 {
		noPredicates := []interface{}{}
		rateLimit := d.Get("rate_limit")

		err := updateWafRateBasedRuleResourceWR(d.Id(), oldPredicates, noPredicates, rateLimit, conn, region)
		if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error updating WAF Regional Rate Based Rule Predicates (%s): %s", d.Id(), err)
		}
	}

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRateBasedRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Regional Rate Based Rule")
		return conn.DeleteRateBasedRule(req)
	})
	if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting WAF Regional Rate Based Rule (%s): %s", d.Id(), err)
	}

	return nil
}

func updateWafRateBasedRuleResourceWR(id string, oldP, newP []interface{}, rateLimit interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRateBasedRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(id),
			Updates:     diffWafRulePredicates(oldP, newP),
			RateLimit:   aws.Int64(int64(rateLimit.(int))),
		}

		return conn.UpdateRateBasedRule(req)
	})

	return err
}
