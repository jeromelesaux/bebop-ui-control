package utils

import (
	"fmt"
	"strings"
)

// hack to remove the right string "_press" from the joystick constant
func RightSuppressButton(name string) string {
	return strings.Replace(name, "_press", "", 1)
}

func RightAddStick(name string) string {
	return fmt.Sprintf("%s_stick", name)
}
