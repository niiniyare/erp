package util

import (
	"strings"
)
func CreateAbbreviation(companyName string) string {
    words := strings.Fields(companyName)
    var abbr strings.Builder
    
    for _, word := range words {
        if len(word) > 0 {
            abbr.WriteString(strings.ToUpper(string(word[0])))
        }
    }
    
    return abbr.String()
}
