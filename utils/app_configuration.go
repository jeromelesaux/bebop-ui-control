package utils

type ApplicationConfig struct {
	Version                string            `json:"version"`
	ActionsJoystickMapping map[string]string `json:"actions_joystick_mapping"`
}

var DroneActions = []string{
	"droneRecording",
	"droneTakeOff",
	"droneStop",
	"droneLand",
	"droneLeftAndRight",
	"droneForwardAndBackward",
	"droneClockwise",
	"droneUpAndDown",
}

var AppConfig *ApplicationConfig = &ApplicationConfig{
	Version: "1.0",
	ActionsJoystickMapping: map[string]string{
		"droneRecording":          "home_press",
		"droneTakeOff":            "start_press",
		"droneStop":               "x_press",
		"droneLand":               "home",
		"droneLeftAndRight":       "left_x",
		"droneForwardAndBackward": "left_y",
		"droneClockwise":          "right_x",
		"droneUpAndDown":          "right_y",
	},
}
