package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// ExpandInvalidSettings is returned when the Expand processor is configured with invalid Input and Output settings.
const ExpandInvalidSettings = errors.Error("ExpandInvalidSettings")

/*
ExpandOptions contains custom options settings for the Expand processor:
	Retain (optional):
		array of JSON keys to retain from the original object
*/
type ExpandOptions struct {
	Retain []string `json:"retain"` // retain fields found anywhere in input
}

/*
Expand processes data by creating individual events from objects in JSON arrays. The processor supports these patterns:
	json array:
		{"expand":[{"foo":"bar"}],"baz":"qux"} >>> {"foo":"bar","baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'expand',
		settings: {
			input_key: 'expand',
			options: {
				retain: ['baz'],
			}
		},
	}
*/
type Expand struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	Options   ExpandOptions            `json:"options"`
}

// Slice processes a slice of bytes with the Expand processor. Conditions are optionally applied on the bytes to enable processing.
func (p Expand) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	// only supports json, error early if there is no input key
	if p.InputKey == "" {
		return nil, fmt.Errorf("slicer settings %v: %v", p, ExpandInvalidSettings)
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		// json array processing
		value := json.Get(data, p.InputKey)
		for _, x := range value.Array() {
			var err error
			processed := []byte(x.String())
			for _, r := range p.Options.Retain {
				v := json.Get(data, r)
				processed, err = json.Set(processed, r, v)
				if err != nil {
					return nil, fmt.Errorf("slicer settings %v: %v", p, err)
				}
			}

			slice = append(slice, processed)
		}
	}

	return slice, nil
}
