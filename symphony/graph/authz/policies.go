// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package authz

import (
	"context"
	"fmt"

	"github.com/facebookincubator/symphony/graph/ent"
	"github.com/facebookincubator/symphony/graph/ent/permissionspolicy"
	"github.com/facebookincubator/symphony/graph/ent/user"
	"github.com/facebookincubator/symphony/graph/ent/usersgroup"
	"github.com/facebookincubator/symphony/graph/graphql/models"
	"github.com/facebookincubator/symphony/graph/viewer"

	models2 "github.com/facebookincubator/symphony/graph/authz/models"
)

// WritePermissionGroupName is the name of the group that its member has write permission for all symphony.
const WritePermissionGroupName = "Write Permission"

var allowedEnums = map[models2.PermissionValue]int{
	models2.PermissionValueNo:          1,
	models2.PermissionValueByCondition: 2,
	models2.PermissionValueYes:         3,
}

func newBasicPermissionRule(allowed bool) *models.BasicPermissionRule {
	rule := models2.PermissionValueNo
	if allowed {
		rule = models2.PermissionValueYes
	}
	return &models.BasicPermissionRule{IsAllowed: rule}
}

func newCUD(allowed bool) *models.Cud {
	return &models.Cud{
		Create: newBasicPermissionRule(allowed),
		Update: newBasicPermissionRule(allowed),
		Delete: newBasicPermissionRule(allowed),
	}
}

func newWorkforceCUD(allowed bool) *models.WorkforceCud {
	return &models.WorkforceCud{
		Create:            newBasicPermissionRule(allowed),
		Update:            newBasicPermissionRule(allowed),
		Delete:            newBasicPermissionRule(allowed),
		Assign:            newBasicPermissionRule(allowed),
		TransferOwnership: newBasicPermissionRule(allowed),
	}
}

// NewInventoryPolicy builds an inventory policy based on general restriction on read,write
func NewInventoryPolicy(readAllowed, writeAllowed bool) *models.InventoryPolicy {
	return &models.InventoryPolicy{
		Read:          newBasicPermissionRule(readAllowed),
		Location:      newCUD(writeAllowed),
		Equipment:     newCUD(writeAllowed),
		EquipmentType: newCUD(writeAllowed),
		LocationType:  newCUD(writeAllowed),
		PortType:      newCUD(writeAllowed),
		ServiceType:   newCUD(writeAllowed),
	}
}

// NewWorkforcePolicy build a workforce policy based on general restriction on read,write
func NewWorkforcePolicy(readAllowed, writeAllowed bool) *models.WorkforcePolicy {
	return &models.WorkforcePolicy{
		Read:      newBasicPermissionRule(readAllowed),
		Data:      newWorkforceCUD(writeAllowed),
		Templates: newCUD(writeAllowed),
	}
}

// NewAdministrativePolicy builds administrative policy of given user
func NewAdministrativePolicy(isAdmin bool) *models.AdministrativePolicy {
	return &models.AdministrativePolicy{
		Access: newBasicPermissionRule(isAdmin),
	}
}

func appendBasicPermissionRule(rule *models.BasicPermissionRule, addRule *models2.BasicPermissionRuleInput) *models.BasicPermissionRule {
	if addRule != nil && allowedEnums[addRule.IsAllowed] >= allowedEnums[rule.IsAllowed] {
		rule.IsAllowed = addRule.IsAllowed
	}
	return rule
}

func appendCUD(cud *models.Cud, addCUD *models2.BasicCUDInput) *models.Cud {
	if addCUD == nil {
		return cud
	}
	cud.Create = appendBasicPermissionRule(cud.Create, addCUD.Create)
	cud.Delete = appendBasicPermissionRule(cud.Delete, addCUD.Delete)
	cud.Update = appendBasicPermissionRule(cud.Update, addCUD.Update)
	return cud
}

func appendWorkforceCUD(cud *models.WorkforceCud, addCUD *models2.BasicWorkforceCUDInput) *models.WorkforceCud {
	if addCUD == nil {
		return cud
	}
	cud.Create = appendBasicPermissionRule(cud.Create, addCUD.Create)
	cud.Delete = appendBasicPermissionRule(cud.Delete, addCUD.Delete)
	cud.Update = appendBasicPermissionRule(cud.Update, addCUD.Update)
	cud.Assign = appendBasicPermissionRule(cud.Assign, addCUD.Assign)
	cud.TransferOwnership = appendBasicPermissionRule(cud.TransferOwnership, addCUD.TransferOwnership)
	return cud
}

// AppendInventoryPolicies append a list of inventory policy inputs to a inventory policy
func AppendInventoryPolicies(policy *models.InventoryPolicy, inputs ...*models2.InventoryPolicyInput) *models.InventoryPolicy {
	for _, input := range inputs {
		if input == nil {
			continue
		}
		policy.Read = appendBasicPermissionRule(policy.Read, input.Read)
		policy.Location = appendCUD(policy.Location, input.Location)
		policy.Equipment = appendCUD(policy.Equipment, input.Equipment)
		policy.EquipmentType = appendCUD(policy.EquipmentType, input.EquipmentType)
		policy.LocationType = appendCUD(policy.LocationType, input.LocationType)
		policy.PortType = appendCUD(policy.PortType, input.PortType)
		policy.ServiceType = appendCUD(policy.ServiceType, input.ServiceType)
	}
	return policy
}

// AppendInventoryPolicies append a list of workforce policy inputs to a workforce policy
func AppendWorkforcePolicies(policy *models.WorkforcePolicy, inputs ...*models2.WorkforcePolicyInput) *models.WorkforcePolicy {
	for _, input := range inputs {
		if input == nil {
			continue
		}
		policy.Read = appendBasicPermissionRule(policy.Read, input.Read)
		policy.Data = appendWorkforceCUD(policy.Data, input.Data)
		policy.Templates = appendCUD(policy.Templates, input.Templates)
	}
	return policy
}

func permissionPolicies(ctx context.Context, v *viewer.UserViewer) (*models.InventoryPolicy, *models.WorkforcePolicy, error) {
	client := ent.FromContext(ctx)
	userID := v.User().ID
	inventoryPolicy := NewInventoryPolicy(false, false)
	workforcePolicy := NewWorkforcePolicy(false, false)
	policies, err := client.PermissionsPolicy.Query().
		Where(permissionspolicy.Or(
			permissionspolicy.IsGlobal(true),
			permissionspolicy.HasGroupsWith(usersgroup.HasMembersWith(user.ID(userID))))).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot query policies: %w", err)
	}
	for _, policy := range policies {
		switch {
		case policy.InventoryPolicy != nil:
			inventoryPolicy = AppendInventoryPolicies(inventoryPolicy, policy.InventoryPolicy)
		case policy.WorkforcePolicy != nil:
			workforcePolicy = AppendWorkforcePolicies(workforcePolicy, policy.WorkforcePolicy)
		default:
			return nil, nil, fmt.Errorf("empty policy found: %d", policy.ID)
		}
	}
	return inventoryPolicy, workforcePolicy, nil
}

func userHasWritePermissions(ctx context.Context) (bool, error) {
	v := viewer.FromContext(ctx)
	if !v.Features().Enabled(viewer.FeatureReadOnly) {
		return true, nil
	}
	if v.Role() == user.RoleOWNER {
		return true, nil
	}
	if v, ok := v.(*viewer.UserViewer); ok {
		return v.User().QueryGroups().
			Where(usersgroup.Name(WritePermissionGroupName)).
			Exist(ctx)
	}
	return false, nil
}

// Permissions builds the aggregated permissions for the given viewer
func Permissions(ctx context.Context) (*models.PermissionSettings, error) {
	writePermissions, err := userHasWritePermissions(ctx)
	if err != nil {
		return nil, err
	}
	v := viewer.FromContext(ctx)
	policiesEnabled := v.Features().Enabled(viewer.FeatureUserManagementDev)
	inventoryPolicy := NewInventoryPolicy(true, writePermissions)
	workforcePolicy := NewWorkforcePolicy(true, writePermissions)
	u, ok := v.(*viewer.UserViewer)
	if !writePermissions && ok && policiesEnabled {
		inventoryPolicy, workforcePolicy, err = permissionPolicies(ctx, u)
		if err != nil {
			return nil, err
		}
	}
	res := models.PermissionSettings{
		// TODO(T64743627): Deprecate CanWrite field
		CanWrite:        writePermissions,
		AdminPolicy:     NewAdministrativePolicy(v.Role() == user.RoleADMIN || v.Role() == user.RoleOWNER),
		InventoryPolicy: inventoryPolicy,
		WorkforcePolicy: workforcePolicy,
	}
	return &res, nil
}

func FullPermissions() *models.PermissionSettings {
	return &models.PermissionSettings{
		CanWrite:        true,
		AdminPolicy:     NewAdministrativePolicy(true),
		InventoryPolicy: NewInventoryPolicy(true, true),
		WorkforcePolicy: NewWorkforcePolicy(true, true),
	}
}

func EmptyPermissions() *models.PermissionSettings {
	return &models.PermissionSettings{
		CanWrite:        false,
		AdminPolicy:     NewAdministrativePolicy(false),
		InventoryPolicy: NewInventoryPolicy(false, false),
		WorkforcePolicy: NewWorkforcePolicy(false, false),
	}
}
