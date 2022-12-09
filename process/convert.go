package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Convert processes data by converting values between types (e.g., string to integer, integer to float). The processor supports these patterns:

	JSON:
		{"convert":"true"} >>> {"convert":true}
		{"convert":"-123"} >>> {"convert":-123}
		{"convert":123} >>> {"convert":"123"}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "convert",
		"settings": {
			"options": {
				"type": "bool"
			},
			"input_key": "convert",
			"output_key": "convert"
		}
	}
*/
type Convert struct {
	Options   ConvertOptions   `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
ConvertOptions contains custom options for the Convert processor:

	Type:
		type that the value is converted to
		must be one of:
			bool (boolean)
			int (integer)
			float
			uint (unsigned integer)
			string
*/
type ConvertOptions struct {
	Type string `json:"type"`
}

// Close closes resources opened by the Convert processor.
func (p Convert) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the Convert processor. Conditions are optionally applied to the data to enable processing.
func (p Convert) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process convert: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process convert: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Convert processor.
func (p Convert) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process convert: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey != "" && p.OutputKey != "" {
		result := capsule.Get(p.InputKey)

		var value interface{}
		switch p.Options.Type {
		case "bool":
			value = result.Bool()
		case "int":
			value = result.Int()
		case "float":
			value = result.Float()
		case "uint":
			value = result.Uint()
		case "string":
			value = result.String()
		}

		if err := capsule.Set(p.OutputKey, value); err != nil {
			return capsule, fmt.Errorf("process convert: %v", err)
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process convert: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}
