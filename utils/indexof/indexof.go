package indexof

// StrIndexof  判断元素是否在字符串数组中
func StrIndexof(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

// IntIndexof  判断元素是否在数值数组中
func IntIndexof(target int64, str_array []int64) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}
