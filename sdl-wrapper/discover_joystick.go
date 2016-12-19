package sdl_wrapper

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"time"
)

var joysticks map[int]*sdl.Joystick = make(map[int]*sdl.Joystick)
var event sdl.Event
var running bool = true

const delay = 1000000000

type JoystickType int

const (
	BUTTON JoystickType = iota
	HAT
	AXIS
)

var ButtonsLabels = []string{
	"x",
	"a",
	"b",
	"y",
	"lb",
	"rb",
	"back",
	"start",
	"home",
	"right_stick",
	"left_stick",
}

var AxisLabels = []string{
	"left_x",
	"left_y",
	"right_x",
	"right_y",
	"rt",
	"lt",
}

// pair is a JSON representation of name and id
type Pair struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// hat is a JSON representation of hat, name and id
type Hat struct {
	Hat  int    `json:"hat"`
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type JoystickConfig struct {
	Name    string `json:"name"`
	GUID    string `json:"guid"`
	Axis    []Pair `json:"axis"`
	Buttons []Pair `json:"buttons"`
	Hats    []Hat  `json:"Hats"`
}

// listen all sdl event from the joystick
// return all informations about the event (for the robots joystick configuration)
func SdlEventData(eventType JoystickType) (which sdl.JoystickID, id uint8, hat uint8, state uint8, sdlType uint32) {
	sdl.Init(sdl.INIT_JOYSTICK)
	sdl.JoystickEventState(sdl.ENABLE)

	defer func() {
		sdl.Quit()
		time.Sleep(delay)
	}()

	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch data := event.(type) {
			case *sdl.JoyAxisEvent:
				if eventType != AXIS {
					continue
				}
				if data.Timestamp < 500 {
					continue
				}
				sdlType = data.Type
				which = data.Which
				id = data.Axis
				fmt.Printf(" [%d ms] JoyAxis\ttype:%d\twhich:%c\taxis:%d\tvalue:%d\n",
					data.Timestamp, data.Type, data.Which, data.Axis, data.Value)

				return
			case *sdl.JoyButtonEvent:
				if eventType != BUTTON {
					continue
				}
				sdlType = data.Type
				which = data.Which
				id = data.Button
				state = data.State
				fmt.Printf(" [%d ms] JoyButton\ttype:%d\twhich:%d\tbutton:%d\tstate:%d\n",
					data.Timestamp, data.Type, data.Which, data.Button, data.State)

				return
			case *sdl.JoyHatEvent:
				if eventType != HAT {
					continue
				}
				sdlType = data.Type
				which = data.Which
				hat = data.Hat
				id = data.Value
				fmt.Printf(" [%d ms] JoyHat\ttype:%d\twhich:%d\that:%d\tvalue:%d\n",
					data.Timestamp, data.Type, data.Which, data.Hat, data.Value)

				return
			case *sdl.JoyDeviceEvent:
				if data.Type == sdl.JOYDEVICEADDED {
					joysticks[int(data.Which)] = sdl.JoystickOpen(data.Which)
				} else if data.Type == sdl.JOYDEVICEREMOVED {
					fmt.Printf("Joystick %d disconnected\n", data.Which)
				}
				continue
			default:
				fmt.Printf("Some event\n")
				continue
			}
		}
	}

	sdl.Quit()
	return
}

func DefineJoystickButtons(jconfig *JoystickConfig) *JoystickConfig {
	for _, button := range ButtonsLabels {
		fmt.Printf("Press %s button.", button)
		_, id, _, _, _ := SdlEventData(BUTTON)
		buttonSetup := Pair{ID: int(id), Name: button}
		jconfig.Buttons = append(jconfig.Buttons, buttonSetup)
	}
	return jconfig
}

func DefineJoystickAxis(jconfig *JoystickConfig) *JoystickConfig {
	for _, axis := range AxisLabels {
		fmt.Printf("Press %s axis.", axis)
		_, id, _, _, _ := SdlEventData(AXIS)
		axisSetup := Pair{ID: int(id), Name: axis}
		jconfig.Axis = append(jconfig.Axis, axisSetup)
	}
	return jconfig
}

func DefineJoystickHats(jconfig *JoystickConfig) *JoystickConfig {
	for _, hat := range AxisLabels {
		fmt.Printf("Press %s hat.", hat)
		_, id, hatValue, _, _ := SdlEventData(HAT)
		hatSetup := Hat{ID: int(id), Name: hat, Hat: int(hatValue)}
		jconfig.Hats = append(jconfig.Hats, hatSetup)
	}
	return jconfig
}
