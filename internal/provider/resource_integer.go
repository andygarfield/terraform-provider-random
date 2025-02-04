package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/random"
)

var _ tfsdk.ResourceType = (*integerResourceType)(nil)

type integerResourceType struct{}

func (r *integerResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "The resource `random_integer` generates random values from a given range, described " +
			"by the `min` and `max` attributes of a given resource.\n" +
			"\n" +
			"This resource can be used in conjunction with resources that have the `create_before_destroy` " +
			"lifecycle flag set, to avoid conflicts with unique names during the brief period where both the " +
			"old and new resources exist concurrently.",
		Attributes: map[string]tfsdk.Attribute{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"min": {
				Description:   "The minimum inclusive value of the range.",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"max": {
				Description:   "The maximum inclusive value of the range.",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"seed": {
				Description:   "A custom seed to always produce the same value.",
				Type:          types.StringType,
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"result": {
				Description: "The random integer result.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"id": {
				Description: "The string representation of the integer result.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (r *integerResourceType) NewResource(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return &integerResource{}, nil
}

var (
	_ tfsdk.Resource                = (*integerResource)(nil)
	_ tfsdk.ResourceWithImportState = (*integerResource)(nil)
)

type integerResource struct{}

func (r *integerResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan integerModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	max := int(plan.Max.Value)
	min := int(plan.Min.Value)
	seed := plan.Seed.Value

	if max < min {
		resp.Diagnostics.AddError(
			"Create Random Integer Error",
			"The minimum (min) value needs to be smaller than or equal to maximum (max) value.",
		)
		return
	}

	rand := random.NewRand(seed)
	number := rand.Intn((max+1)-min) + min

	u := &integerModelV0{
		ID:      types.String{Value: strconv.Itoa(number)},
		Keepers: plan.Keepers,
		Min:     types.Int64{Value: int64(min)},
		Max:     types.Int64{Value: int64(max)},
		Result:  types.Int64{Value: int64(number)},
	}

	if seed != "" {
		u.Seed.Value = seed
	} else {
		u.Seed.Null = true
	}

	diags = resp.State.Set(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *integerResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r *integerResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *integerResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

func (r *integerResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	parts := strings.Split(req.ID, ",")
	if len(parts) != 3 && len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Import Random Integer Error",
			"Invalid import usage: expecting {result},{min},{max} or {result},{min},{max},{seed}",
		)
		return
	}

	result, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Random Integer Error",
			"The value supplied could not be parsed as an integer.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	min, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Random Integer Error",
			"The min value supplied could not be parsed as an integer.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	max, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Random Integer Error",
			"The max value supplied could not be parsed as an integer.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var state integerModelV0

	state.ID.Value = parts[0]
	state.Keepers.ElemType = types.StringType
	state.Result.Value = result
	state.Min.Value = min
	state.Max.Value = max

	if len(parts) == 4 {
		state.Seed.Value = parts[3]
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type integerModelV0 struct {
	ID      types.String `tfsdk:"id"`
	Keepers types.Map    `tfsdk:"keepers"`
	Min     types.Int64  `tfsdk:"min"`
	Max     types.Int64  `tfsdk:"max"`
	Seed    types.String `tfsdk:"seed"`
	Result  types.Int64  `tfsdk:"result"`
}
