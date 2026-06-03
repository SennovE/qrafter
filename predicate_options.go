package qrafter

// PredicateOption is an optional predicate used for compact dynamic filters.
type PredicateOption struct {
	predicate Predicater
}

// Predicates returns only predicates enabled by the given options.
func Predicates(options ...PredicateOption) []Predicater {
	predicates := make([]Predicater, 0, len(options))
	for _, option := range options {
		if option.predicate != nil {
			predicates = append(predicates, option.predicate)
		}
	}
	return predicates
}

// When includes predicate when condition is true.
func When(condition bool, predicate Predicater) PredicateOption {
	if !condition {
		return PredicateOption{}
	}
	return PredicateOption{predicate: predicate}
}

// WhenFunc builds and includes a predicate lazily when condition is true.
func WhenFunc(condition bool, build func() Predicater) PredicateOption {
	if !condition {
		return PredicateOption{}
	}
	return PredicateOption{predicate: build()}
}

// WhenPtr builds and includes a predicate when value is not nil.
func WhenPtr[T any](value *T, build func(T) Predicater) PredicateOption {
	if value == nil {
		return PredicateOption{}
	}
	return PredicateOption{predicate: build(*value)}
}

// WhenNotEmpty builds and includes a predicate when value is not empty.
func WhenNotEmpty(value string, build func(string) Predicater) PredicateOption {
	if value == "" {
		return PredicateOption{}
	}
	return PredicateOption{predicate: build(value)}
}
