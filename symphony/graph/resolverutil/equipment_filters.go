// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package resolverutil

import (
	"github.com/facebookincubator/symphony/graph/ent"
	"github.com/facebookincubator/symphony/graph/ent/equipment"
	"github.com/facebookincubator/symphony/graph/ent/equipmenttype"
	"github.com/facebookincubator/symphony/graph/ent/property"
	"github.com/facebookincubator/symphony/graph/ent/propertytype"
	"github.com/facebookincubator/symphony/graph/graphql/models"

	"github.com/pkg/errors"
)

func handleEquipmentFilter(q *ent.EquipmentQuery, filter *models.EquipmentFilterInput) (*ent.EquipmentQuery, error) {
	switch filter.FilterType {
	case models.EquipmentFilterTypeEquipInstName:
		return equipmentNameFilter(q, filter)
	case models.EquipmentFilterTypeProperty:
		return equipmentPropertyFilter(q, filter)
	}
	return nil, errors.Errorf("filter type is not supported: %s", filter.FilterType)
}

func equipmentNameFilter(q *ent.EquipmentQuery, filter *models.EquipmentFilterInput) (*ent.EquipmentQuery, error) {
	if filter.Operator == models.FilterOperatorContains {
		return q.Where(equipment.NameContainsFold(*filter.StringValue)), nil
	}
	return nil, errors.Errorf("operation %q not supported", filter.Operator)
}

func equipmentPropertyFilter(q *ent.EquipmentQuery, filter *models.EquipmentFilterInput) (*ent.EquipmentQuery, error) {
	p := filter.PropertyValue
	switch filter.Operator {
	case models.FilterOperatorIs:
		q = q.Where(
			equipment.HasPropertiesWith(
				property.HasTypeWith(
					propertytype.Name(p.Name),
					propertytype.Type(p.Type.String()),
				),
			),
		)
		pred, err := GetPropertyPredicate(*p)
		if err != nil {
			return nil, err
		}
		if pred != nil {
			q = q.Where(equipment.HasPropertiesWith(pred))
		}
		return q, nil
	case models.FilterOperatorDateLessThan, models.FilterOperatorDateGreaterThan:
		if p.Type != models.PropertyKindDate {
			return nil, errors.Errorf("property kind should be type")
		}

		propPred := property.StringValGT(*p.StringValue)
		propTypePred := propertytype.StringValGT(*p.StringValue)
		if filter.Operator == models.FilterOperatorDateLessThan {
			propPred = property.StringValLT(*p.StringValue)
			propTypePred = propertytype.StringValLT(*p.StringValue)
		}
		q = q.Where(equipment.Or(
			equipment.HasPropertiesWith(
				property.And(
					property.HasTypeWith(
						propertytype.Name(p.Name),
						propertytype.Type(p.Type.String()),
					),
					propPred,
				),
			),
			equipment.And(
				equipment.HasTypeWith(equipmenttype.HasPropertyTypesWith(
					propertytype.Name(p.Name),
					propertytype.Type(p.Type.String()),
					propTypePred,
				)),
				equipment.Not(equipment.HasPropertiesWith(
					property.HasTypeWith(
						propertytype.Name(p.Name),
						propertytype.Type(p.Type.String()),
					)),
				))))
		return q, nil
	default:
		return nil, errors.Errorf("operator %q not supported", filter.Operator)
	}
}