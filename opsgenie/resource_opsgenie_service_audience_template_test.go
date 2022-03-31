package opsgenie

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ogClient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/service"
)

func init() {
	resource.AddTestSweepers("opsgenie_service_audience_template", &resource.Sweeper{
		Name: "opsgenie_service_audience_template",
		F:    testSweepServiceAudienceTemplate,
	})

}

func testSweepServiceAudienceTemplate(region string) error {
	meta, err := sharedConfigForRegion()
	if err != nil {
		return err
	}

	client, err := service.NewClient(meta.(*OpsgenieClient).client.Config)
	if err != nil {
		return err
	}
	resp, err := client.List(context.Background(), &service.ListRequest{})
	if err != nil {
		return err
	}

	for _, svc := range resp.Services {
		if strings.HasPrefix(svc.Name, "genietest-") {
			log.Printf("Destroying service %s", svc.Name)

			deleteRequest := service.DeleteRequest{
				Id: svc.Id,
			}

			if _, err := client.Delete(context.Background(), &deleteRequest); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccOpsGenieServiceAudienceTemplate_basic(t *testing.T) {
	randomTeam := acctest.RandString(6)
	randomResponder := acctest.RandString(6)
	randomService := acctest.RandString(6)

	config := testAccOpsGenieServiceAudienceTemplate_basic(randomTeam, randomResponder, randomService)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      testCheckOpsGenieServiceAudienceTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckOpsGenieServiceAudienceTemplateExists("opsgenie_service_audience_template.test"),
				),
			},
		},
	})
}

func testCheckOpsGenieServiceAudienceTemplateDestroy(s *terraform.State) error {
	client, err := service.NewClient(testAccProvider.Meta().(*OpsgenieClient).client.Config)
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opsgenie_service" {
			continue
		}
		req := service.GetRequest{
			Id: rs.Primary.Attributes["id"],
		}
		_, err := client.Get(context.Background(), &req)
		if err != nil {
			x := err.(*ogClient.ApiError)
			if x.StatusCode != 404 {
				return errors.New(fmt.Sprintf("Service still exists : %s", x.Error()))
			}
		}
	}

	return nil
}

func testCheckOpsGenieServiceAudienceTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		service_id := rs.Primary.Attributes["service_id"]

		client, err := service.NewClient(testAccProvider.Meta().(*OpsgenieClient).client.Config)
		if err != nil {
			return err
		}
		_, err = client.GetAudienceTemplate(context.Background(), &service.GetAudienceTemplateRequest{
			ServiceId: service_id,
		})

		if err != nil {
			return fmt.Errorf("Bad: Service ID %q does not exist", service_id)
		}

		return nil
	}
}

func testAccOpsGenieServiceAudienceTemplate_basic(randomTeam, randomResponder, randomService string) string {
	return fmt.Sprintf(`
	resource "opsgenie_team" "test" {
		name        = "genieteam-%s"
		description = "This team deals with all the things"
	}
	resource "opsgenie_team" "test-responder" {
		name        = "genieteam-%s"
		description = "This team is a responder"
	}
	resource "opsgenie_service" "test" {
		name  = "genietest-service-%s"
		team_id = opsgenie_team.test.id
	}
	resource "opsgenie_service_audience_template" "test" {
		service_id = opsgenie_service.test.id
		audience_template {
			responder {
				teams = [opsgenie_team.test-responder.id]
			}
		}
	}
	`, randomTeam, randomResponder, randomService)
}
