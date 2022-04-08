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
		Create: resourceOpsGenieServiceAudienceTemplateUpdate,
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
				Type:     schema.TypeSet,
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

	audience_template, err := client.GetAudienceTemplate(context.Background(), &service.GetAudienceTemplateRequest{
		ServiceId: service_id,
	})
	if err != nil {
		return err
	}

	d.Set("service_id", service_id)
	// audience_teamplate := make()
	d.Set("audience_template", flattenOpsgenieServiceAudienceTemplate(audience_template))
	// d.Set("stakeholder", result.Stakeholder)

	return nil
}

func resourceOpsGenieServiceAudienceTemplateUpdate(input *schema.ResourceData, meta interface{}) error {
	client, err := service.NewClient(meta.(*OpsgenieClient).client.Config)
	if err != nil {
		return err
	}

	service_id := input.Get("service_id").(string)
	updateRequest := &service.UpdateAudienceTemplateRequest{
		ServiceId: service_id,
	}

	audience_template := input.Get("audience_template").(*schema.Set)

	if audience_template != nil {
		for _, v := range audience_template.List() {
			config := v.(map[string]interface{})

			responder := config["responder"].(*schema.Set)
			if len(responder.List()) > 0 {
				updateRequest.Responder = expandOpsGenieServiceAudienceTemplateResponder(responder)
			}

			stakeholder := config["stakeholder"].(*schema.Set)
			if len(stakeholder.List()) > 0 {
				updateRequest.Stakeholder = expandOpsGenieServiceAudienceTemplateStakeholder(stakeholder)
			}
		}
	}

	log.Printf("[INFO] Updating OpsGenie Service Audience Template for service '%s'", input.Get("service_id").(string))
	_, err = client.UpdateAudienceTemplate(context.Background(), updateRequest)
	if err != nil {
		return err
	}

	return resourceOpsGenieServiceAudienceTemplateRead(input, meta)
}

func resourceOpsGenieServiceAudienceTemplateDelete(input *schema.ResourceData, meta interface{}) error {
	client, err := service.NewClient(meta.(*OpsgenieClient).client.Config)
	if err != nil {
		return err
	}

	// delete updates with nil values
	service_id := input.Get("service_id").(string)
	updateRequest := &service.UpdateAudienceTemplateRequest{
		ServiceId:   service_id,
		Responder:   service.ResponderOfAudience{},
		Stakeholder: service.StakeholderOfAudience{},
	}

	log.Printf("[INFO] Deleting OpsGenie Service Audience Template for service '%s'", input.Get("service_id").(string))
	_, err = client.UpdateAudienceTemplate(context.Background(), updateRequest)
	if err != nil {
		return err
	}

	return nil
}

func expandOpsGenieServiceAudienceTemplateResponder(input *schema.Set) service.ResponderOfAudience {
	responder := service.ResponderOfAudience{}
	if input == nil {
		return responder
	}

	for _, v := range input.List() {
		config := v.(map[string]interface{})

		teams := config["teams"].(*schema.Set)
		if len(teams.List()) > 0 {
			responder.Teams = flattenOpsgenieServiceAudienceTemplateRequestTeams(teams)
		}

		individuals := config["individuals"].(*schema.Set)
		if len(individuals.List()) > 0 {
			responder.Individuals = flattenOpsgenieServiceAudienceTemplateRequestIndividuals(config["individuals"].(*schema.Set))
		}
	}
	return responder
}

func expandOpsGenieServiceAudienceTemplateStakeholder(input *schema.Set) service.StakeholderOfAudience {
	StakeholderOfAudience := service.StakeholderOfAudience{}
	if input != nil {
		for _, v := range input.List() {
			config := v.(map[string]interface{})

			if config["individuals"].(*schema.Set).Len() > 0 {
				StakeholderOfAudience.Individuals = flattenOpsgenieServiceAudienceTemplateRequestIndividuals(config["individuals"].(*schema.Set))
			}
			conditionMatchType := config["condition_match_type"].(string)
			if len(conditionMatchType) > 0 {
				StakeholderOfAudience.ConditionMatchType = og.ConditionMatchType(conditionMatchType)
			}
			conditions := config["conditions"].([]interface{})
			if len(conditions) > 0 {
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

func flattenOpsgenieServiceAudienceTemplate(input *service.GetAudienceTemplateResult) []map[string]interface{} {
	template := make([]map[string]interface{}, 0, 1)
	out := make(map[string]interface{})
	out["responder"] = flattenOpsgenieServiceAudienceTemplateResponder(input.Responder)
	out["stakeholder"] = flattenOpsgenieServiceAudienceTemplateStakeholder(input.Stakeholder)
	template = append(template, out)
	return template
}

func flattenOpsgenieServiceAudienceTemplateStakeholder(input service.StakeholderOfAudience) []map[string]interface{} {
	stakeholder := make([]map[string]interface{}, 0, 1)
	out := make(map[string]interface{})
	if len(input.Conditions) > 0{
		out["conditions"] = input.Conditions
	}
	if len(input.ConditionMatchType) > 0 {
		out["individuals"] = input.Individuals
	}
	stakeholder = append(stakeholder, out)
	return stakeholder	
}

func flattenOpsgenieServiceAudienceTemplateResponder(input service.ResponderOfAudience) []map[string]interface{} {
	responder := make([]map[string]interface{}, 0, 1)
	out := make(map[string]interface{})
	if len(input.Teams) > 0{
		out["teams"] = input.Teams
	}
	if len(input.Individuals) > 0 {
		out["individuals"] = input.Individuals
	}
	responder = append(responder, out)
	return responder
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
