package opsgenie

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opsgenie/opsgenie-go-sdk-v2/og"
	"github.com/opsgenie/opsgenie-go-sdk-v2/service"
)

func resourceOpsGenieServiceAudienceTemplate() *schema.Resource {
	return &schema.Resource{
		Read:   handleNonExistentResource(resourceOpsGenieServiceAudienceTemplateRead),
		Update: resourceOpsGenieServiceAudienceTemplateUpdate,
		Delete: resourceOpsGenieServiceAudienceTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 130),
			},
			"audience_template": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"responder": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"teams": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 50,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"individuals": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 50,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"stakeholder": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"individuals": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 50,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"condition_match_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "match-any-condition",
										ValidateFunc: validation.StringInSlice([]string{"match-all-conditions", "match-any-condition"}, false),
									},
									"conditions": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"match_field": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														"country", "state", "city",
														"zipcode", "line", "tag", "custom-property",
													}, false),
												},
												"key": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Required if match_field is custom-property.",
												},
												"value": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Value to be checked for the match field",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceOpsGenieServiceAudienceTemplateRead(d *schema.ResourceData, meta interface{}) error {
	client, err := service.NewClient(meta.(*OpsgenieClient).client.Config)
	if err != nil {
		return err
	}
	service_id := d.Get("service_id").(string)

	log.Printf("[INFO] Reading OpsGenie Service Audience Template for service: '%s'", service_id)

	result, err := client.GetAudienceTemplate(context.Background(), &service.GetAudienceTemplateRequest{
		ServiceId: service_id,
	})
	if err != nil {
		return err
	}

	d.Set("service_id", service_id)
	d.Set("responder", result.Responder)
	d.Set("stakeholder", result.Stakeholder)

	return nil
}

func resourceOpsGenieServiceAudienceTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := service.NewClient(meta.(*OpsgenieClient).client.Config)
	if err != nil {
		return err
	}

	service_id := d.Get("service_id").(string)
	updateRequest := &service.UpdateAudienceTemplateRequest{
		ServiceId: service_id,
	}

	audience_template := d.Get("audience_template").([]interface{})
	for _, v := range audience_template {
		config := v.(map[string]interface{})
		updateRequest.Responder = expandOpsGenieServiceAudienceTemplateResponder(config["responder"].([]interface{}))
		updateRequest.Stakeholder = expandOpsGenieServiceAudienceTemplateStakeholder(config["stakeholder"].([]interface{}))
	}

	log.Printf("[INFO] Updating OpsGenie Service Audience Template for service '%s'", d.Get("service_id").(string))
	_, err = client.UpdateAudienceTemplate(context.Background(), updateRequest)
	if err != nil {
		return err
	}

	return resourceOpsGenieServiceAudienceTemplateRead(d, meta)
}

func resourceOpsGenieServiceAudienceTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := service.NewClient(meta.(*OpsgenieClient).client.Config)
	if err != nil {
		return err
	}

	// delete updates with nil values
	service_id := d.Get("service_id").(string)
	updateRequest := &service.UpdateAudienceTemplateRequest{
		ServiceId:   service_id,
		Responder:   service.ResponderOfAudience{},
		Stakeholder: service.StakeholderOfAudience{},
	}

	log.Printf("[INFO] Deleting OpsGenie Service Audience Template for service '%s'", d.Get("service_id").(string))
	_, err = client.UpdateAudienceTemplate(context.Background(), updateRequest)
	if err != nil {
		return err
	}

	return nil
}

func expandOpsGenieServiceAudienceTemplateResponder(input []interface{}) service.ResponderOfAudience {
	ResponderOfAudience := service.ResponderOfAudience{}
	if input != nil {
		for _, v := range input {
			config := v.(map[string]interface{})

			if config["teams"].(*schema.Set).Len() > 0 {
				ResponderOfAudience.Teams = flattenOpsgenieServiceAudienceTemplateRequestTeams(config["teams"].(*schema.Set))
			}

			if config["individuals"].(*schema.Set).Len() > 0 {
				ResponderOfAudience.Individuals = flattenOpsgenieServiceAudienceTemplateRequestIndividuals(config["individuals"].(*schema.Set))

			}
		}
	}

	return ResponderOfAudience
}

func expandOpsGenieServiceAudienceTemplateStakeholder(input []interface{}) service.StakeholderOfAudience {
	StakeholderOfAudience := service.StakeholderOfAudience{}
	if input != nil {
		for _, v := range input {
			config := v.(map[string]interface{})

			if config["individuals"].(*schema.Set).Len() > 0 {
				StakeholderOfAudience.Individuals = flattenOpsgenieServiceAudienceTemplateRequestIndividuals(config["individuals"].(*schema.Set))
			}
			if config["condition_match_type"].(*schema.Set).Len() > 0 {
				StakeholderOfAudience.ConditionMatchType = og.ConditionMatchType(config["condition_match_type"].(string))
			}
			if config["conditions"].(*schema.Set).Len() > 0 {
				StakeholderOfAudience.Conditions = expandOpsGenieServiceAudienceTemplateConditions(config["conditions"].([]interface{}))
			}
		}
	}
	return StakeholderOfAudience
}

func expandOpsGenieServiceAudienceTemplateConditions(input []interface{}) []service.ConditionOfStakeholder {
	conditions := make([]service.ConditionOfStakeholder, 0, len(input))
	if input == nil {
		return conditions
	}

	for _, v := range input {
		condition := service.ConditionOfStakeholder{}
		config := v.(map[string]interface{})
		condition.MatchField = service.MatchField(config["match_field"].(string))
		condition.Value = config["value"].(string)
		if condition.MatchField == service.CustomProperty {
			key := config["key"].(string)
			if key != "" {
				condition.Key = config["key"].(string)
			}
		}
		conditions = append(conditions, condition)
	}

	return conditions
}

func flattenOpsgenieServiceAudienceTemplateRequestIndividuals(input *schema.Set) []string {
	individual := make([]string, len(input.List()))
	if input == nil {
		return individual
	}

	for k, v := range input.List() {
		individual[k] = v.(string)
	}
	return individual
}

func flattenOpsgenieServiceAudienceTemplateRequestTeams(input *schema.Set) []string {
	team := make([]string, len(input.List()))
	if input == nil {
		return team
	}

	for k, v := range input.List() {
		team[k] = v.(string)
	}
	return team
}
