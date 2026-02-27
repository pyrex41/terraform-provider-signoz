package resource

import "testing"

func TestNormalizeJSON(t *testing.T) {
	t.Run("compacts and sorts object keys", func(t *testing.T) {
		in := "{\n  \"b\": 2,\n  \"a\": 1\n}"
		got := normalizeJSON(in)
		want := "{\"a\":1,\"b\":2}"

		if got != want {
			t.Fatalf("normalizeJSON() mismatch\nwant: %s\ngot:  %s", want, got)
		}
	})

	t.Run("unescapes HTML entities", func(t *testing.T) {
		in := "{\"expr\":\"a \\u003e= b \\u0026\\u0026 c \\u003c d\"}"
		got := normalizeJSON(in)
		want := "{\"expr\":\"a >= b && c < d\"}"

		if got != want {
			t.Fatalf("normalizeJSON() mismatch\nwant: %s\ngot:  %s", want, got)
		}
	})

	t.Run("returns input for invalid JSON", func(t *testing.T) {
		in := "{not-json}"
		got := normalizeJSON(in)

		if got != in {
			t.Fatalf("normalizeJSON() should return original string for invalid JSON\nwant: %s\ngot:  %s", in, got)
		}
	})
}

func TestSemanticallyEqualJSON(t *testing.T) {
	t.Run("equivalent JSON with key reordering", func(t *testing.T) {
		a := "{\"query\":{\"builder\":{\"queryData\":[{\"filter\":{\"expression\":\"deployment.environment IN $deployment.environment\"},\"aggregations\":[{\"metricName\":\"signoz_latency.count\",\"timeAggregation\":\"rate\",\"spaceAggregation\":\"sum\"}]}]}}}"
		b := "{\"query\":{\"builder\":{\"queryData\":[{\"aggregations\":[{\"spaceAggregation\":\"sum\",\"timeAggregation\":\"rate\",\"metricName\":\"signoz_latency.count\"}],\"filter\":{\"expression\":\"deployment.environment IN $deployment.environment\"}}]}}}"

		if !semanticallyEqualJSON(a, b) {
			t.Fatalf("expected semanticallyEqualJSON() to return true for equivalent JSON")
		}
	})

	t.Run("different array ordering is not equal", func(t *testing.T) {
		a := "[1,2,3]"
		b := "[3,2,1]"

		if semanticallyEqualJSON(a, b) {
			t.Fatalf("expected semanticallyEqualJSON() to return false for different array ordering")
		}
	})

	t.Run("different filter operators are not equal", func(t *testing.T) {
		a := "{\"filters\":{\"items\":[{\"op\":\"in\",\"value\":[\"$deployment.environment\"]}]}}"
		b := "{\"filters\":{\"items\":[{\"op\":\"=\",\"value\":\"$deployment.environment\"}]}}"

		if semanticallyEqualJSON(a, b) {
			t.Fatalf("expected semanticallyEqualJSON() to return false for behavior-changing JSON")
		}
	})

	t.Run("invalid JSON falls back to strict string compare", func(t *testing.T) {
		a := "{not-json}"
		b := "{not-json}"
		c := "{still-not-json}"

		if !semanticallyEqualJSON(a, b) {
			t.Fatalf("expected identical invalid JSON strings to be equal")
		}
		if semanticallyEqualJSON(a, c) {
			t.Fatalf("expected different invalid JSON strings to be different")
		}
	})
}
