package resources

func Collect[T any](items []T, transform func(T) map[string]any) []map[string]any {
	out := make([]map[string]any, len(items))
	for i, item := range items {
		out[i] = transform(item)
	}
	return out
}
