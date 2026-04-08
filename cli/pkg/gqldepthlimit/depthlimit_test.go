package gqldepthlimit_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"

	"hmans.de/chatto/pkg/gqldepthlimit"
)

func field(name string, children ...ast.Selection) *ast.Field {
	return &ast.Field{
		Alias: name,
		Name:  name,
		SelectionSet: ast.SelectionSet(children),
	}
}

func TestSelectionSetDepth_EmptySet(t *testing.T) {
	depth := gqldepthlimit.SelectionSetDepth(nil, nil, 0)
	if depth != 0 {
		t.Errorf("empty selection set: got depth %d, want 0", depth)
	}
}

func TestSelectionSetDepth_FlatFields(t *testing.T) {
	// { id name }  → depth 1
	set := ast.SelectionSet{field("id"), field("name")}
	depth := gqldepthlimit.SelectionSetDepth(set, nil, 0)
	if depth != 1 {
		t.Errorf("flat fields: got depth %d, want 1", depth)
	}
}

func TestSelectionSetDepth_NestedFields(t *testing.T) {
	// { user { profile { name } } }  → depth 3
	set := ast.SelectionSet{
		field("user",
			field("profile",
				field("name"),
			),
		),
	}
	depth := gqldepthlimit.SelectionSetDepth(set, nil, 0)
	if depth != 3 {
		t.Errorf("nested fields: got depth %d, want 3", depth)
	}
}

func TestSelectionSetDepth_MaxBranch(t *testing.T) {
	// { a { b } c { d { e } } }  → depth 3 (via c.d.e)
	set := ast.SelectionSet{
		field("a", field("b")),
		field("c", field("d", field("e"))),
	}
	depth := gqldepthlimit.SelectionSetDepth(set, nil, 0)
	if depth != 3 {
		t.Errorf("max branch: got depth %d, want 3", depth)
	}
}

func TestSelectionSetDepth_IntrospectionExempt(t *testing.T) {
	// { __schema { types { name } } user { id } }  → depth 1 (only user.id counts)
	set := ast.SelectionSet{
		field("__schema",
			field("types",
				field("name"),
			),
		),
		field("user", field("id")),
	}
	depth := gqldepthlimit.SelectionSetDepth(set, nil, 0)
	if depth != 2 {
		t.Errorf("introspection exempt: got depth %d, want 2", depth)
	}
}

func TestSelectionSetDepth_InlineFragmentNoDepth(t *testing.T) {
	// { user { ... on User { name } } }  → depth 2 (inline fragment doesn't add)
	set := ast.SelectionSet{
		field("user",
			&ast.InlineFragment{
				SelectionSet: ast.SelectionSet{field("name")},
			},
		),
	}
	depth := gqldepthlimit.SelectionSetDepth(set, nil, 0)
	if depth != 2 {
		t.Errorf("inline fragment: got depth %d, want 2", depth)
	}
}

func TestSelectionSetDepth_FragmentSpread(t *testing.T) {
	// { user { ...UserFields } }
	// fragment UserFields on User { profile { name } }
	// → depth 3 (user > profile > name)
	fragments := ast.FragmentDefinitionList{
		{
			Name: "UserFields",
			SelectionSet: ast.SelectionSet{
				field("profile", field("name")),
			},
		},
	}

	set := ast.SelectionSet{
		field("user",
			&ast.FragmentSpread{Name: "UserFields"},
		),
	}
	depth := gqldepthlimit.SelectionSetDepth(set, fragments, 0)
	if depth != 3 {
		t.Errorf("fragment spread: got depth %d, want 3", depth)
	}
}

func TestSelectionSetDepth_UnknownFragmentIgnored(t *testing.T) {
	set := ast.SelectionSet{
		field("user",
			&ast.FragmentSpread{Name: "NonExistent"},
		),
	}
	depth := gqldepthlimit.SelectionSetDepth(set, nil, 0)
	if depth != 1 {
		t.Errorf("unknown fragment: got depth %d, want 1", depth)
	}
}

func TestExtensionName(t *testing.T) {
	ext := &gqldepthlimit.Extension{MaxDepth: 10}
	if name := ext.ExtensionName(); name != "QueryDepthLimit" {
		t.Errorf("ExtensionName() = %q, want %q", name, "QueryDepthLimit")
	}
}

func TestValidate(t *testing.T) {
	ext := &gqldepthlimit.Extension{MaxDepth: 10}
	if err := ext.Validate(nil); err != nil {
		t.Errorf("Validate() returned unexpected error: %v", err)
	}
}

// makeOpCtx builds a minimal OperationContext with the given selection set and fragments.
func makeOpCtx(set ast.SelectionSet, fragments ast.FragmentDefinitionList) *graphql.OperationContext {
	return &graphql.OperationContext{
		Doc: &ast.QueryDocument{
			Fragments: fragments,
		},
		Operation: &ast.OperationDefinition{
			SelectionSet: set,
		},
	}
}

func TestMutateOperationContext_UnderLimit(t *testing.T) {
	// { user { name } } → depth 2, limit 5 → should pass
	ext := &gqldepthlimit.Extension{MaxDepth: 5}
	opCtx := makeOpCtx(ast.SelectionSet{field("user", field("name"))}, nil)

	err := ext.MutateOperationContext(context.Background(), opCtx)
	if err != nil {
		t.Errorf("expected nil error for depth under limit, got: %v", err)
	}
}

func TestMutateOperationContext_AtLimit(t *testing.T) {
	// { a { b { c } } } → depth 3, limit 3 → should pass (not exceeding)
	ext := &gqldepthlimit.Extension{MaxDepth: 3}
	opCtx := makeOpCtx(ast.SelectionSet{
		field("a", field("b", field("c"))),
	}, nil)

	err := ext.MutateOperationContext(context.Background(), opCtx)
	if err != nil {
		t.Errorf("expected nil error at exact limit, got: %v", err)
	}
}

func TestMutateOperationContext_OverLimit(t *testing.T) {
	// { a { b { c { d } } } } → depth 4, limit 3 → should reject
	ext := &gqldepthlimit.Extension{MaxDepth: 3}
	opCtx := makeOpCtx(ast.SelectionSet{
		field("a", field("b", field("c", field("d")))),
	}, nil)

	err := ext.MutateOperationContext(context.Background(), opCtx)
	if err == nil {
		t.Fatal("expected error for depth exceeding limit, got nil")
	}

	// Verify error message contains depth info
	if got := err.Message; got == "" {
		t.Error("error message should not be empty")
	}

	// Verify error code is set
	found := false
	for _, ext := range err.Extensions {
		if ext == "QUERY_DEPTH_LIMIT_EXCEEDED" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("error extensions should contain QUERY_DEPTH_LIMIT_EXCEEDED, got: %v", err.Extensions)
	}
}

func TestMutateOperationContext_OverLimit_WithFragments(t *testing.T) {
	// { user { ...DeepFields } }
	// fragment DeepFields on User { a { b { c } } }
	// → depth 4 (user > a > b > c), limit 3 → should reject
	ext := &gqldepthlimit.Extension{MaxDepth: 3}
	fragments := ast.FragmentDefinitionList{
		{
			Name:         "DeepFields",
			SelectionSet: ast.SelectionSet{field("a", field("b", field("c")))},
		},
	}
	opCtx := makeOpCtx(ast.SelectionSet{
		field("user", &ast.FragmentSpread{Name: "DeepFields"}),
	}, fragments)

	err := ext.MutateOperationContext(context.Background(), opCtx)
	if err == nil {
		t.Fatal("expected error for fragment-based depth exceeding limit, got nil")
	}
}

func TestMutateOperationContext_IntrospectionExempt(t *testing.T) {
	// { __schema { types { name { deep { deeper } } } } } → introspection, exempt
	// Even though this is deeply nested, introspection fields are exempt
	ext := &gqldepthlimit.Extension{MaxDepth: 1}
	opCtx := makeOpCtx(ast.SelectionSet{
		field("__schema", field("types", field("name", field("deep", field("deeper"))))),
	}, nil)

	err := ext.MutateOperationContext(context.Background(), opCtx)
	if err != nil {
		t.Errorf("introspection should be exempt from depth limit, got: %v", err)
	}
}

func TestMutateOperationContext_EmptyQuery(t *testing.T) {
	ext := &gqldepthlimit.Extension{MaxDepth: 5}
	opCtx := makeOpCtx(nil, nil)

	err := ext.MutateOperationContext(context.Background(), opCtx)
	if err != nil {
		t.Errorf("empty query should pass, got: %v", err)
	}
}
