package iam

// ruleToValues pads a rule slice to exactly 6 string slots.
func ruleToValues(rule []string) (v0, v1, v2, v3, v4, v5 string) {
	padded := make([]string, 6)
	copy(padded, rule)
	return padded[0], padded[1], padded[2], padded[3], padded[4], padded[5]
}

// filterEmpty removes trailing empty strings from a slice.
func filterEmpty(ss []string) []string {
	last := len(ss) - 1
	for last >= 0 && ss[last] == "" {
		last--
	}
	return ss[:last+1]
}

// joinRule joins rule values with ", " for persist.LoadPolicyLine.
func joinRule(rule []string) string {
	out := ""
	for i, v := range rule {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}
