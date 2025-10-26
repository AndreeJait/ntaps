package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndreeJait/ntaps/internal/paths"
	"github.com/AndreeJait/ntaps/internal/util"
)

// ensureRequestDTOHasPathParams makes sure <MethodName>Request includes fields for each path param.
// e.g. {transaction_code} -> TransactionCode string `param:"transaction_code"`
func ensureRequestDTOHasPathParams(
	ucPkg string,
	ucMethodName string,
	pathParams []string,
) error {

	dtoPath := filepath.Join(paths.RootUsecaseDir, ucPkg, "dto.go")

	raw, err := os.ReadFile(dtoPath)
	if err != nil {
		return err
	}
	src := string(raw)

	typeName := ucMethodName + "Request"

	// if dto.go doesn't have the Request type at all, we can't safely edit automatically
	typeDeclStart := strings.Index(src, "type "+typeName+" struct")
	if typeDeclStart == -1 {
		return nil
	}

	// find the closing "}" of that struct
	braceOpen := strings.Index(src[typeDeclStart:], "{")
	if braceOpen == -1 {
		return nil
	}
	braceOpen += typeDeclStart

	braceClose := strings.Index(src[braceOpen:], "}")
	if braceClose == -1 {
		return nil
	}
	braceClose += braceOpen

	structBody := src[braceOpen:braceClose] // includes leading "{\n..."

	updatedBody := structBody

	for _, p := range pathParams {
		fieldName := util.ToPascalCase(p) // transaction_code -> TransactionCode
		fieldLine := fmt.Sprintf("\n\t%s string `param:\"%s\"`", fieldName, p)

		// only add if not present
		if !strings.Contains(updatedBody, fieldName+" ") {
			updatedBody = updatedBody + fieldLine
		}
	}

	// rebuild full file content
	src = src[:braceOpen] + updatedBody + src[braceClose:]

	return util.WriteGoFile(dtoPath, src)
}
