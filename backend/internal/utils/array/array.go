package array

func FindFirst[T any](in []T, find func(T) bool) (int, T, bool) {
	for i, t := range in {
		if find(t) {
			return i, t, true
		}
	}
	return -1, *new(T), false
}

func FindIndex[T any](in []T, find func(T) bool) (int, bool) {
	for index, t := range in {
		if find(t) {
			return index, true
		}
	}
	return -1, false
}

func RemoveItem[T any](in []T, find func(T) bool) []T {
	index := -1
	for i, t := range in {
		if find(t) {
			index = i
			break
		}
	}
	if index == -1 {
		return in
	}

	return append(in[:index], in[index+1:]...)
}

func RemoveDuplicates[T any](in []T, find func(T, T) bool) []T {
	list := []T{}
	for _, item := range in {
		_, _, ok := FindFirst(list, func(c T) bool {
			return find(c, item)
		})
		if !ok {
			list = append(list, item)
		}
	}
	return list
}

func IsEqual[T any](a []T, b []T, find func(T, T) bool) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		_, _, ok := FindFirst(b, func(c T) bool {
			return find(c, v)
		})
		if !ok {
			return false
		}
	}
	return true
}

func Contains[T any](contain, contained []T, find func(T, T) bool) bool {
	if len(contain) < len(contained) || len(contain) == 0 {
		return false
	}
	for _, v := range contained {
		_, _, ok := FindFirst(contain, func(c T) bool {
			return find(c, v)
		})
		if !ok {
			return false
		}
	}
	return true
}

func Filter[T any](in []T, filter func(T) bool) []T {
	list := []T{}
	for _, item := range in {
		if filter(item) {
			list = append(list, item)
		}
	}
	return list
}
