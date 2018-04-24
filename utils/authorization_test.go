// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package utils

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/model"
)

type RoleState struct {
	RoleName   string `json:"roleName"`
	Permission string `json:"permission"`
	ShouldHave bool   `json:"shouldHave"`
}

func mockConfig() *model.Config {
	config := model.Config{}
	config.SetDefaults()
	return &config
}

func mapping() (map[string]map[string][]RoleState, error) {

	policiesRolesMapping := make(map[string]map[string][]RoleState)

	raw, err := ioutil.ReadFile("./policies-roles-mapping.json")
	if err != nil {
		return policiesRolesMapping, err
	}

	var f map[string]interface{}
	err = jsoniter.Unmarshal(raw, &f)
	if err != nil {
		return policiesRolesMapping, err
	}

	for policyName, value := range f {

		capitalizedName := fmt.Sprintf("%v%v", strings.ToUpper(policyName[:1]), policyName[1:])
		policiesRolesMapping[capitalizedName] = make(map[string][]RoleState)

		for policyValue, roleStatesMappings := range value.(map[string]interface{}) {

			var roleStates []RoleState
			for _, roleStateMapping := range roleStatesMappings.([]interface{}) {

				roleStateMappingJSON, _ := jsoniter.Marshal(roleStateMapping)
				var roleState RoleState
				_ = jsoniter.Unmarshal(roleStateMappingJSON, &roleState)

				roleStates = append(roleStates, roleState)

			}

			policiesRolesMapping[capitalizedName][policyValue] = roleStates

		}

	}

	return policiesRolesMapping, nil
}

func TestSetRolePermissionsFromConfig(t *testing.T) {

	mapping, err := mapping()
	if err != nil {
		require.NoError(t, err)
	}

	for policyName, v := range mapping {
		for policyValue, rolesMappings := range v {

			config := mockConfig()
			updateConfig(config, policyName, policyValue)
			roles := model.MakeDefaultRoles()
			SetRolePermissionsFromConfig(roles, config, true)

			for _, roleMappingItem := range rolesMappings {
				role := roles[roleMappingItem.RoleName]

				permission := roleMappingItem.Permission
				hasPermission := roleHasPermission(role, permission)

				if (roleMappingItem.ShouldHave && !hasPermission) || (!roleMappingItem.ShouldHave && hasPermission) {
					wording := "not to"
					if roleMappingItem.ShouldHave {
						wording = "to"
					}
					t.Errorf("Expected '%v' %v have '%v' permission when '%v' is set to '%v'.", role.Name, wording, permission, policyName, policyValue)
				}

			}

		}
	}
}

func updateConfig(config *model.Config, key string, value string) {
	v := reflect.ValueOf(config.ServiceSettings)
	field := v.FieldByName(key)
	if !field.IsValid() {
		v = reflect.ValueOf(config.TeamSettings)
		field = v.FieldByName(key)
	}

	switch value {
	case "true", "false":
		b, _ := strconv.ParseBool(value)
		field.Elem().SetBool(b)
	default:
		field.Elem().SetString(value)
	}
}

func roleHasPermission(role *model.Role, permission string) bool {
	for _, p := range role.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
