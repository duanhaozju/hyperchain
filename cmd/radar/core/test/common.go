package test

func Judge(result1 map[string][]string, result2 map[string][]string) bool {
	if len(result1) != len(result2) {
		return false
	}
	for k1, v1 := range result1 {
		if v2, ok := result2[k1]; ok {
			if len(v1) != len(v2) {
				return false
			}
			for i := 0; i < len(v1); i++ {
				flag := false
				for j := 0; j < len(v2); j++ {
					if v1[i] == v2[j] {
						v2 = append(v2[:j], v2[j+1:]...)
						flag = true
						break
					}
				}
				if !flag {
					return false
				}
			}
			if len(v2) != 0 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
