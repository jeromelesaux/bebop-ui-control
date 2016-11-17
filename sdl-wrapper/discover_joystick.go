package sdl_wrapper

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
)

var joysticks map[int]*sdl.Joystick = make(map[int]*sdl.Joystick)
var event sdl.Event
var running bool = true

type JoystickType int

const (
	BUTTON JoystickType = iota
	HAT
	AXIS
)

// listen all sdl event from the joystick
// return all informations about the event (for the robots joystick configuration)
func SdlEventData(eventType JoystickType) (which sdl.JoystickID, id uint8, hat uint8, state uint8, sdlType uint32) {
	sdl.Init(sdl.INIT_JOYSTICK)
	sdl.JoystickEventState(sdl.ENABLE)

	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch data := event.(type) {
			case *sdl.JoyAxisEvent:
				if eventType != AXIS {
					continue
				}
				sdlType = data.Type
				which = data.Which
				id = data.Axis
				fmt.Printf("JoyAxisEvent:%d,%d\n", which, id)
				sdl.Quit()
				return
			case *sdl.JoyButtonEvent:
				if eventType != BUTTON {
					continue
				}
				sdlType = data.Type
				which = data.Which
				id = data.Button
				state = data.State
				fmt.Printf("JoyButtonEvent:%d,%d,%d\n", which, id, state)
				sdl.Quit()
				return
			case *sdl.JoyHatEvent:
				if eventType != HAT {
					continue
				}
				sdlType = data.Type
				which = data.Which
				hat = data.Hat
				id = data.Value
				fmt.Printf("JoyHatEvent:%d,%d,%d\n", which, id, hat)
				sdl.Quit()
				return
			//case *sdl.KeyDownEvent:
			//	fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c\tmodifiers:%d\tstate:%d\trepeat:%d\n",
			//		data.Timestamp, data.Type, data.Keysym.Sym, data.Keysym.Mod, data.State, data.Repeat)
			//	continue
			case *sdl.JoyDeviceEvent:
				if data.Type == sdl.JOYDEVICEADDED {
					joysticks[int(data.Which)] = sdl.JoystickOpen(data.Which)
					//if joysticks[int(data.Which)] != nil {
					//	fmt.Printf("Joystick %d connected\n", data.Which)
					//}
				} else if data.Type == sdl.JOYDEVICEREMOVED {
					//if joystick := joysticks[int(data.Which)]; joystick != nil {
					//	// remove the joystick from slice
					//	joystick.Close()
					//}

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
