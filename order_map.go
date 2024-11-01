package waitprocess

// orderMap is a map with order, not thread safe
type orderMap[K comparable, V any] struct {
	data  map[K]V
	order []K
}

func newOrderMap[K comparable, V any]() *orderMap[K, V] {
	return &orderMap[K, V]{
		data:  make(map[K]V),
		order: make([]K, 0),
	}
}

func (om *orderMap[K, V]) size() int {
	return len(om.data)
}

// set set key value, if key exists, update value and keep order
func (om *orderMap[K, V]) set(key K, value V) {
	if om.contains(key) {
		om.data[key] = value
		return
	}

	om.data[key] = value
	om.order = append(om.order, key)
}

func (om *orderMap[K, V]) contains(key K) bool {
	_, ok := om.data[key]
	return ok
}

func (om *orderMap[K, V]) load(key K) (bool, V) {
	value, ok := om.data[key]
	return ok, value
}

func (om *orderMap[K, V]) get(key K) V {
	value := om.data[key]
	return value
}

func (om *orderMap[K, V]) delete(key K) bool {
	if _, ok := om.data[key]; !ok {
		return false
	}

	delete(om.data, key)
	for i, k := range om.order {
		if k == key {
			om.order = append(om.order[:i], om.order[i+1:]...)
			break
		}
	}

	return true
}

func (om *orderMap[K, V]) rangeFunc(f func(index int, key K, value V) bool) {
	for i, key := range om.order {
		value := om.data[key]
		if !f(i, key, value) {
			break
		}
	}
}
