---
layout: "opsgenie"
page_title: "Opsgenie: opsgenie_service_audience_template"
sidebar_current: "docs-opsgenie-resource-service-audience-template"
description: |-
  Manages a Service Audience Template within Opsgenie.
---

# opsgenie\_service\_audience\_template

Manages a Service Audience Template within Opsgenie.

## Example Usage

```hcl
resource "opsgenie_team" "test" {
  name        = "example-team"
  description = "This team deals with all the things"
}

resource "opsgenie_team" "test-responder" {
    name        = "example-responder"
    description = "This team is a responder"
}

resource "opsgenie_service" "test" {
    name  = "example-service"
    team_id = opsgenie_team.test.id
}

resource "opsgenie_user" "test" {
    username  = "stakeholder@example.com"
    full_name = "Stake Holder"
    role      = "User"
}

resource "opsgenie_service_audience_template" "test" {
    service_id = opsgenie_service.test.id
    audience_template {
        responder {
            teams = [opsgenie_team.test-responder.id]
        }
        stakeholder {
            individuals = [opsgenie_user.test.id]
        }
    }
}
```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required) ID of the service associated

* `audience_template` - (Required) This is the configuration for the service audience template. This is a block, structure is documented below.

The `audience_template` block supports:

* `responder` - (Optional) Responders of an incident. This is a block, structure is documented below. 

* `stakeholder` - (Optional) Stakeholders of an incident. This is a block, structure is documented below. 


The `responder` block supports:

* `teams` - (Optional) List of team ids to be used as responder teams

* `individuals` - (Optional) List of user ids to be used as responder users.


The `stakeholder` block supports:

* `individuals` - (Optional) List of user ids to be used as stakeholders.

* `conditionMatchType` - (Optional) Match type for given conditions. Possible values are `match-all-conditions`,  `match-any-condition`. Default value is `match-any-condition`.

* `conditions` - (Optional) Condition list which contains match-type conditions. This is a block, structure is documented below. 


The `conditions` block supports:

* `matchField` - (Required) Field to be matched for users. Possible values are `country`, `state`, `city`, `zipCode`, `line`, `tag`, `customProperty`.

* `key` - (Optional) If matchField is `customProperty`, key must be given.

* `value` - (Required) Description that is generally used to provide a detailed information about the alert.


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Opsgenie Service.

## Import

Service Audience Template can be imported using the `service_id`, e.g.

`$ terraform import opsgenie_service.this service_id`
