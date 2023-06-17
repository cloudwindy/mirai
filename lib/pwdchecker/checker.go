package pwdchecker

import (
	"errors"
	"fmt"
	"strings"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	replaceChars      = `!@$&*`
	sepChars          = `_-., `
	otherSpecialChars = `"#%'()+/:;<=>?[\]^{|}~`
	lowerChars        = `abcdefghijklmnopqrstuvwxyz`
	upperChars        = `ABCDEFGHIJKLMNOPQRSTUVWXYZ`
	digitsChars       = `0123456789`
)

// Validate returns nil if the password has greater than or
// equal to the minimum entropy. If not, an error is returned
// that explains how the password can be strengthened. This error
// is safe to show the client
func Validate(password string, minEntropy float64) error {
	entropy := passwordvalidator.GetEntropy(password)
	if entropy >= minEntropy {
		return nil
	}

	hasReplace := false
	hasSep := false
	hasOtherSpecial := false
	hasLower := false
	hasUpper := false
	hasDigits := false
	for _, c := range password {
		if strings.ContainsRune(replaceChars, c) {
			hasReplace = true
			continue
		}
		if strings.ContainsRune(sepChars, c) {
			hasSep = true
			continue
		}
		if strings.ContainsRune(otherSpecialChars, c) {
			hasOtherSpecial = true
			continue
		}
		if strings.ContainsRune(lowerChars, c) {
			hasLower = true
			continue
		}
		if strings.ContainsRune(upperChars, c) {
			hasUpper = true
			continue
		}
		if strings.ContainsRune(digitsChars, c) {
			hasDigits = true
			continue
		}
	}

	allMessages := []string{}

	if !hasOtherSpecial || !hasSep || !hasReplace {
		allMessages = append(allMessages, "特殊符号")
	}
	if !hasLower {
		allMessages = append(allMessages, "小写字符")
	}
	if !hasUpper {
		allMessages = append(allMessages, "大写字符")
	}
	if !hasDigits {
		allMessages = append(allMessages, "数字")
	}

	if len(allMessages) > 0 {
		return fmt.Errorf(
			"密码过于简单，请尝试添加%v。当前密码复杂度: %.2f (最低要求: %.2f)",
			strings.Join(allMessages, "、"),
			entropy,
			minEntropy,
		)
	}

	return errors.New("密码过于简单，请增加密码长度。")
}
