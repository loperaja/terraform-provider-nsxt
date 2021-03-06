/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	"strings"
)

func dataSourceNsxtPolicyTier0Gateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyTier0GatewayRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"edge_cluster_path": {
				Type:        schema.TypeString,
				Description: "The path of the edge cluster connected to this Tier0 gateway",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicyTier0GatewayReadAllTier0(connector *client.RestConnector) ([]model.Tier0, error) {
	var results []model.Tier0
	client := infra.NewDefaultTier0sClient(connector)
	boolFalse := false
	var cursor *string
	total := 0

	for {
		gateways, err := client.List(cursor, &boolFalse, nil, nil, &boolFalse, nil)
		if err != nil {
			return results, err
		}
		results = append(results, gateways.Results...)
		if total == 0 && gateways.ResultCount != nil {
			// first response
			total = int(*gateways.ResultCount)
		}
		cursor = gateways.Cursor
		if len(results) >= total {
			return results, nil
		}
	}
}

func dataSourceNsxtPolicyTier0GatewayRead(d *schema.ResourceData, m interface{}) error {
	// Read a tier0 by name or id
	connector := getPolicyConnector(m)
	client := infra.NewDefaultTier0sClient(connector)

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj model.Tier0
	if objID != "" {
		// Get by id
		objGet, err := client.Get(objID)

		if err != nil {
			return handleDataSourceReadError(d, "Tier0", objID, err)
		}
		obj = objGet
	} else if objName == "" {
		return fmt.Errorf("Error obtaining Tier0 ID or name during read")
	} else {
		// Get by full name/prefix
		objList, err := dataSourceNsxtPolicyTier0GatewayReadAllTier0(connector)
		if err != nil {
			return handleListError("Tier0", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.Tier0
		var prefixMatch []model.Tier0
		for _, objInList := range objList {
			if strings.HasPrefix(*objInList.DisplayName, objName) {
				prefixMatch = append(prefixMatch, objInList)
			}
			if *objInList.DisplayName == objName {
				perfectMatch = append(perfectMatch, objInList)
			}
		}
		if len(perfectMatch) > 0 {
			if len(perfectMatch) > 1 {
				return fmt.Errorf("Found multiple Tier0s with name '%s'", objName)
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("Found multiple Tier0s with name starting with '%s'", objName)
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("Tier0 router '%s' was not found", objName)
		}
	}

	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)
	return nil
}
