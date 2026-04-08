// Package gqldepthlimit provides a gqlgen extension that rejects GraphQL
// operations exceeding a configurable nesting depth. This prevents malicious
// queries that exploit circular type references to cause excessive resource
// consumption.
package gqldepthlimit

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const errCode = "QUERY_DEPTH_LIMIT_EXCEEDED"

// Extension rejects GraphQL operations that exceed the maximum nesting depth.
type Extension struct {
	MaxDepth int
}

var _ interface {
	graphql.OperationContextMutator
	graphql.HandlerExtension
} = &Extension{}

func (d *Extension) ExtensionName() string {
	return "QueryDepthLimit"
}

func (d *Extension) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (d *Extension) MutateOperationContext(_ context.Context, opCtx *graphql.OperationContext) *gqlerror.Error {
	depth := SelectionSetDepth(opCtx.Operation.SelectionSet, opCtx.Doc.Fragments, 0)
	if depth > d.MaxDepth {
		err := gqlerror.Errorf("operation has depth %d, which exceeds the limit of %d", depth, d.MaxDepth)
		errcode.Set(err, errCode)
		return err
	}
	return nil
}

// SelectionSetDepth calculates the maximum nesting depth of a selection set,
// resolving fragment spreads and inline fragments.
//
// Introspection fields (__schema, __type, __typename) are exempt from depth
// counting — they query the schema itself, not user data, and can't exploit
// circular type references.
//
// Inline fragments don't add depth — they're type-narrowing, not nesting.
func SelectionSetDepth(set ast.SelectionSet, fragments ast.FragmentDefinitionList, currentDepth int) int {
	if len(set) == 0 {
		return currentDepth
	}

	maxDepth := currentDepth
	for _, sel := range set {
		var childDepth int
		switch s := sel.(type) {
		case *ast.Field:
			if strings.HasPrefix(s.Name, "__") {
				continue
			}
			childDepth = SelectionSetDepth(s.SelectionSet, fragments, currentDepth+1)
		case *ast.InlineFragment:
			childDepth = SelectionSetDepth(s.SelectionSet, fragments, currentDepth)
		case *ast.FragmentSpread:
			frag := fragments.ForName(s.Name)
			if frag != nil {
				childDepth = SelectionSetDepth(frag.SelectionSet, fragments, currentDepth)
			}
		}
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}
	return maxDepth
}
