package forum

import (
	"fmt"
	// "regexp"
)

func Ascii(str string) error {
    for _, char := range str {
        if char > 127 {
            return fmt.Errorf("error: String contains non-ASCII characters")
        }
    }
    return nil
}


