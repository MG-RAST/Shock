package node

import (
	"testing"
)

func TestSanitizeQuery_AllowsSafeOperators(t *testing.T) {
	tests := []struct {
		name  string
		query map[string]interface{}
	}{
		{
			name:  "simple field match",
			query: map[string]interface{}{"file.name": "test.fasta"},
		},
		{
			name:  "$gt operator",
			query: map[string]interface{}{"file.size": map[string]interface{}{"$gt": 1000}},
		},
		{
			name:  "$in operator",
			query: map[string]interface{}{"file.name": map[string]interface{}{"$in": []interface{}{"a", "b"}}},
		},
		{
			name:  "$exists operator",
			query: map[string]interface{}{"attributes.project": map[string]interface{}{"$exists": true}},
		},
		{
			name: "$and with $elemMatch",
			query: map[string]interface{}{
				"$and": []interface{}{
					map[string]interface{}{"file.size": map[string]interface{}{"$gt": 0}},
					map[string]interface{}{"tags": map[string]interface{}{"$elemMatch": map[string]interface{}{"$eq": "metagenome"}}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sanitizeQuery(tt.query); err != nil {
				t.Errorf("sanitizeQuery() returned unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizeQuery_BlocksDangerousOperators(t *testing.T) {
	tests := []struct {
		name     string
		query    map[string]interface{}
		wantOp   string
	}{
		{
			name:   "$where at top level",
			query:  map[string]interface{}{"$where": "this.file.size > 1000"},
			wantOp: "$where",
		},
		{
			name:   "$expr at top level",
			query:  map[string]interface{}{"$expr": map[string]interface{}{"$gt": []interface{}{"$file.size", 1000}}},
			wantOp: "$expr",
		},
		{
			name:   "$function at top level",
			query:  map[string]interface{}{"$function": map[string]interface{}{"body": "function() { return true; }"}},
			wantOp: "$function",
		},
		{
			name:   "$accumulator at top level",
			query:  map[string]interface{}{"$accumulator": map[string]interface{}{"init": "function() {}"}},
			wantOp: "$accumulator",
		},
		{
			name: "$where nested inside $or",
			query: map[string]interface{}{
				"$or": []interface{}{
					map[string]interface{}{"file.name": "test"},
					map[string]interface{}{"$where": "this.file.size > 0"},
				},
			},
			wantOp: "$where",
		},
		{
			name: "$expr nested inside $and inside $or",
			query: map[string]interface{}{
				"$or": []interface{}{
					map[string]interface{}{
						"$and": []interface{}{
							map[string]interface{}{"$expr": map[string]interface{}{"$gt": []interface{}{"$a", "$b"}}},
						},
					},
				},
			},
			wantOp: "$expr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizeQuery(tt.query)
			if err == nil {
				t.Errorf("sanitizeQuery() expected error for operator %s, got nil", tt.wantOp)
				return
			}
			expected := "operator " + tt.wantOp + " is not allowed in passthrough queries"
			if err.Error() != expected {
				t.Errorf("sanitizeQuery() error = %q, want %q", err.Error(), expected)
			}
		})
	}
}
