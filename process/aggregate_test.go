package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var aggregateTests = []struct {
	name     string
	proc     Aggregate
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"single aggregate",
		Aggregate{
			Options: AggregateOptions{
				Separator: `\n`,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}\n{"baz":"qux"}\n{"quux":"corge"}`),
		},
		nil,
	},
	// identical to the single buffer test, but improves performance due to lowered count and size limits
	{
		"single aggregate with better performance",
		Aggregate{
			Options: AggregateOptions{
				Separator: `\n`,
				MaxCount:  3,
				MaxSize:   50,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}\n{"baz":"qux"}\n{"quux":"corge"}`),
		},
		nil,
	},
	{
		"multiple aggregates",
		Aggregate{
			Options: AggregateOptions{
				Separator: `\n`,
				MaxCount:  2,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}\n{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		nil,
	},
	{
		"single JSON array aggregate",
		Aggregate{
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"aggregate":[{"foo":"bar"},{"baz":"qux"},{"quux":"corge"}]}`),
			[]byte(`{"aggregate":[{"fofo":"bar"},{"baz":"qux"},{"quux":"corge"}]}`),
		},
		nil,
	},
	{
		"multiple JSON array aggregates",
		Aggregate{
			Options: AggregateOptions{
				MaxCount: 2,
			},
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"aggregate":[{"foo":"bar"},{"baz":"qux"}]}`),
			[]byte(`{"aggregate":[{"quux":"corge"}]}`),
		},
		nil,
	},
	{
		"single JSON array aggregate",
		Aggregate{
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`foo`),
			[]byte(`bar`),
			[]byte(`baz`),
			[]byte(`qux`),
			[]byte(`quux`),
			[]byte(`corge`),
		},
		[][]byte{
			[]byte(`{"aggregate":["foo","bar","baz","qux","quux","corge"]}`),
		},
		nil,
	},
	{
		"multiple JSON array aggregates",
		Aggregate{
			Options: AggregateOptions{
				MaxCount: 2,
			},
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`foo`),
			[]byte(`bar`),
			[]byte(`baz`),
			[]byte(`qux`),
			[]byte(`quux`),
			[]byte(`corge`),
		},
		[][]byte{
			[]byte(`{"aggregate":["foo","bar"]}`),
			[]byte(`{"aggregate":["baz","qux"]}`),
			[]byte(`{"aggregate":["quux","corge"]}`),
		},
		nil,
	},
	{
		"JSON key aggregates",
		Aggregate{
			Options: AggregateOptions{
				AggregateKey: "foo",
			},
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"qux"}`),
			[]byte(`{"foo":"bar"}`),
		},
		[][]byte{
			[]byte(`{"aggregate":[{"foo":"bar"},{"foo":"bar"},{"foo":"bar"}]}`),
			[]byte(`{"aggregate":[{"foo":"baz"}]}`),
			[]byte(`{"aggregate":[{"foo":"qux"}]}`),
		},
		nil,
	},
}

func TestAggregate(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range aggregateTests {
		var capsules []config.Capsule
		for _, t := range test.test {
			capsule.SetData(t)
			capsules = append(capsules, capsule)
		}

		result, err := test.proc.ApplyBatch(ctx, capsules)
		if err != nil {
			t.Error(err)
		}

		for i, res := range result {
			expected := test.expected[i]
			if !bytes.Equal(expected, res.Data()) {
				t.Errorf("expected %s, got %s", expected, string(res.Data()))
			}
		}
	}
}

func benchmarkAggregate(b *testing.B, batcher Aggregate, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = batcher.ApplyBatch(ctx, capsules)
	}
}

func BenchmarkAggregate(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range aggregateTests {
		b.Run(test.name,
			func(b *testing.B) {
				var capsules []config.Capsule
				for _, t := range test.test {
					_ = capsule.SetData(t)
					capsules = append(capsules, capsule)
				}

				benchmarkAggregate(b, test.proc, capsules)
			},
		)
	}
}
