package types

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// DaysOfWeekType is a custom type for days_of_the_week that implements semantic equality
// to ignore element ordering differences when comparing API responses.
type DaysOfWeekType struct {
	basetypes.ListType
}

// Equal returns true if the given type is equal to this type.
func (t DaysOfWeekType) Equal(o attr.Type) bool {
	other, ok := o.(DaysOfWeekType)
	if !ok {
		return false
	}
	return t.ListType.Equal(other.ListType)
}

// String returns a human-readable string representation of the type.
func (t DaysOfWeekType) String() string {
	return "DaysOfWeekType"
}

// ValueFromList converts a basetypes.ListValue to a DaysOfWeekValue.
func (t DaysOfWeekType) ValueFromList(ctx context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	value := DaysOfWeekValue{
		ListValue: in,
	}

	return value, nil
}

// ValueFromTerraform converts a tftypes.Value to a DaysOfWeekValue.
func (t DaysOfWeekType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ListType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	listValue, ok := attrValue.(basetypes.ListValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type: expected basetypes.ListValue, got %T", attrValue)
	}

	listValuable, diags := t.ValueFromList(ctx, listValue)
	if diags.HasError() {
		return nil, fmt.Errorf("error converting to DaysOfWeekValue: %v", diags)
	}

	return listValuable, nil
}

// ValueType returns the value type for this custom type.
func (t DaysOfWeekType) ValueType(ctx context.Context) attr.Value {
	return DaysOfWeekValue{}
}

// TerraformType returns the tftypes.Type that represents this type.
func (t DaysOfWeekType) TerraformType(ctx context.Context) tftypes.Type {
	return t.ListType.TerraformType(ctx)
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the type.
func (t DaysOfWeekType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return t.ListType.ApplyTerraform5AttributePathStep(step)
}

// DaysOfWeekValue is a custom value type that implements semantic equality for days_of_the_week.
// It compares lists as unordered sets to prevent false drift detection when the API
// returns days in a different order than configured.
type DaysOfWeekValue struct {
	basetypes.ListValue
}

// Type returns the type of this value.
func (v DaysOfWeekValue) Type(ctx context.Context) attr.Type {
	return DaysOfWeekType{
		ListType: basetypes.ListType{ElemType: v.ListValue.ElementType(ctx)},
	}
}

// Equal returns true if the given value is equal to this value.
// This performs strict equality checking (type and value must match exactly).
func (v DaysOfWeekValue) Equal(o attr.Value) bool {
	other, ok := o.(DaysOfWeekValue)
	if !ok {
		return false
	}
	return v.ListValue.Equal(other.ListValue)
}

// ListSemanticEquals implements semantic equality for days_of_the_week.
// Returns true if both lists contain the same days, regardless of order.
// This prevents false drift detection when the API returns days in different order.
func (v DaysOfWeekValue) ListSemanticEquals(ctx context.Context, other basetypes.ListValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Get the other value as DaysOfWeekValue
	otherValue, ok := other.(DaysOfWeekValue)
	if !ok {
		return false, diags
	}

	// If both are null, they're equal
	if v.IsNull() && otherValue.IsNull() {
		return true, diags
	}

	// If one is null and the other isn't, they're not equal
	if v.IsNull() != otherValue.IsNull() {
		return false, diags
	}

	// If both are unknown, they're equal
	if v.IsUnknown() && otherValue.IsUnknown() {
		return true, diags
	}

	// If one is unknown and the other isn't, they're not equal
	if v.IsUnknown() != otherValue.IsUnknown() {
		return false, diags
	}

	// Extract elements from both lists
	var thisElements, otherElements []attr.Value
	diags.Append(v.ElementsAs(ctx, &thisElements, false)...)
	if diags.HasError() {
		return false, diags
	}

	diags.Append(otherValue.ElementsAs(ctx, &otherElements, false)...)
	if diags.HasError() {
		return false, diags
	}

	// If lengths differ, they're not equal
	if len(thisElements) != len(otherElements) {
		return false, diags
	}

	// Convert to []int64 and sort both for comparison
	thisInts := make([]int64, len(thisElements))
	for i, elem := range thisElements {
		intVal, ok := elem.(basetypes.Int64Value)
		if !ok {
			diags.AddError(
				"Invalid Element Type",
				fmt.Sprintf("Expected Int64Value, got %T", elem),
			)
			return false, diags
		}
		thisInts[i] = intVal.ValueInt64()
	}

	otherInts := make([]int64, len(otherElements))
	for i, elem := range otherElements {
		intVal, ok := elem.(basetypes.Int64Value)
		if !ok {
			diags.AddError(
				"Invalid Element Type",
				fmt.Sprintf("Expected Int64Value, got %T", elem),
			)
			return false, diags
		}
		otherInts[i] = intVal.ValueInt64()
	}

	// Sort both slices
	sort.Slice(thisInts, func(i, j int) bool { return thisInts[i] < thisInts[j] })
	sort.Slice(otherInts, func(i, j int) bool { return otherInts[i] < otherInts[j] })

	// Compare sorted slices
	for i := range thisInts {
		if thisInts[i] != otherInts[i] {
			return false, diags
		}
	}

	// Lists contain same elements (order-independent)
	return true, diags
}
