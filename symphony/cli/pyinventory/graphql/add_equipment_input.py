#!/usr/bin/env python3
# @generated AUTOGENERATED file. Do not Change!

from dataclasses import dataclass
from datetime import datetime
from functools import partial
from gql.gql.datetime_utils import DATETIME_FIELD
from numbers import Number
from typing import Any, Callable, List, Mapping, Optional

from dataclasses_json import DataClassJsonMixin

from .property_input import PropertyInput
@dataclass
class AddEquipmentInput(DataClassJsonMixin):
    name: str
    type: str
    properties: List[PropertyInput]
    location: Optional[str] = None
    parent: Optional[str] = None
    positionDefinition: Optional[str] = None
    workOrder: Optional[str] = None
    externalId: Optional[str] = None

