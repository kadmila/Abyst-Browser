package functional

func Filter[T any, Q any](s []T, f func(T) Q) []Q {
	result := make([]Q, len(s))
	for i, e := range s {
		result[i] = f(e)
	}
	return result
}

func Filter_ok[T any, Q any](s []T, f func(T) (Q, bool)) []Q {
	result := make([]Q, 0, len(s))
	for _, e := range s {
		v, ok := f(e)
		if ok {
			result = append(result, v)
		}
	}
	return result
}

func Filter_strict_ok[T any, Q any](s []T, f func(T) (Q, bool)) ([]Q, bool) {
	result := make([]Q, len(s))
	for i, e := range s {
		v, ok := f(e)
		if !ok {
			return nil, false
		}
		result[i] = v
	}
	return result, true
}

func Filter_until_err[T any, Q any](s []T, f func(T) (Q, error)) ([]Q, []T, error) {
	result := make([]Q, 0, len(s))
	for i, e := range s {
		v, err := f(e)
		if err != nil {
			return result, s[i:], err
		}
		result = append(result, v)
	}
	return result, nil, nil
}
