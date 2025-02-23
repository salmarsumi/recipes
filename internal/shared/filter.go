package shared

// Filter filters elements of the input slice based on a predicate function.
// It returns a new slice containing only the elements that satisfy the predicate.
//
// T represents any type.
//
// Parameters:
//   - input: A slice of elements of type T to be filtered.
//   - predicate: A function that takes an element of type T and returns a boolean.
//                The element is included in the result if the predicate returns true.
//  - selector: A function that takes an element of type T and returns a result of type TResult.
// 			    It is used to transform each element of the input slice that satisfies the predicate.
//
// Returns:
//   A slice of elements of type TResult that satisfy the predicate function.
func Filter[T any, TResult any](input []T, predicate func(T) bool, selector func(T) TResult) (result []TResult) {
	for _, value := range input {
		if predicate(value) {
			result = append(result, selector(value))
		}
	}
	return result
}
