package docs

// Spec returns the complete OpenAPI 3.0 JSON specification.
// Split into two constants to keep source files manageable.
func Spec() string {
return specPart1 + specPart2
}