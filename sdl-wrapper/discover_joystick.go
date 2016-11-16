package sdl_wrapper

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
)

// listen all sdl event from the joystick
// return all informations about the event (for the robots joystick configuration)
func SdlEventData() (which sdl.JoystickID, id uint8, hat uint8, state uint8, sdlType uint32) {
	for {
		event := sdl.PollEvent()
		switch data := event.(type) {
		case *sdl.JoyAxisEvent:
			sdlType = data.Type
			which = data.Which
			id = data.Axis
			fmt.Printf("JoyAxisEvent:%d,%d\n", which, id)
			return
		case *sdl.JoyButtonEvent:
			sdlType = data.Type
			which = data.Which
			id = data.Button
			state = data.State
			fmt.Printf("JoyButtonEvent:%d,%d,%d\n", which, id, state)
			return
		case *sdl.JoyHatEvent:
			sdlType = data.Type
			which = data.Which
			hat = data.Hat
			id = data.Value
			fmt.Printf("JoyHatEvent:%d,%d,%d\n", which, id, hat)
			return
		}
	}

}
