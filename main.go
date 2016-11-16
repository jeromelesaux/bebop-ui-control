package main

import (
	"bebop-ui-control/sdl-wrapper"
	"bebop-ui-control/utils"
	"flag"
	"fmt"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/bebop"
	"github.com/hybridgroup/gobot/platforms/joystick"
	"github.com/hybridgroup/gobot/platforms/keyboard"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"time"
)

// arguments used for the app
// discover joystick's button, hat and axis from sdl layer
var testjoystick = flag.Bool("testjoystick", false, "test your joystick control id, false by default")

// overload the joystick configuration file path
var jConfig = flag.String("joystickconfig", "", "Path of the joystick config")

// global varirables
var recording = false
var Logger *log.Logger
var drone *bebop.BebopDriver
var rightStick = pair{x: 0, y: 0}
var leftStick = pair{x: 0, y: 0}
var offset = 32767.0

func main() {
	// default joystick file path
	joystickConfigPath := "joystick_control_conf.json"
	// parse all the arguments
	flag.Parse()

	// append or create the log file
	flog, err := os.OpenFile("bebop-ui-control.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		fmt.Fprintf(os.Stdout, "Error while creating log file "+err.Error())
	}
	defer flog.Close()

	// manage the joystick  configuration file path
	if *jConfig != "" {
		joystickConfigPath = *jConfig
	}

	Logger = log.New(flog, "Message : ", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Println("Logger service initiated.")

	// create a new gobot
	gbot := gobot.NewGobot()
	gbot.AutoStop = true

	// create default joystick adaptor
	joystickAdaptor := joystick.NewJoystickAdaptor("ps3")
	Logger.Println("Joystick adapter service initiated.")
	stick := joystick.NewJoystickDriver(joystickAdaptor,
		"ps3",
		joystickConfigPath,
	)

	// if discover mode set, only execute discoverJoystick function
	// press key Q to exit the program
	if *testjoystick == true {
		Logger.Println("Discover service started.")
		discoverJoystick(gbot, stick)
	}

	// create drone driver adaptor
	bebopAdaptor := bebop.NewBebopAdaptor("Drone")
	Logger.Println("Bebop adapter service initiated.")
	drone = bebop.NewBebopDriver(bebopAdaptor, "Drone")
	Logger.Println("Bebop driver service initiated.")
	Logger.Println("Drone and gobots services initiated.")

	// create a default keyboard drive
	// press key Q to exit the program
	keys := keyboard.NewKeyboardDriver("keyboard")

	droneWork := func() {

		// define the drone video return and use a new thread
		video, _, _ := ffmpeg()

		go func() {
			for {
				if _, err := video.Write(<-drone.Video()); err != nil {
					fmt.Println(err)
					return
				}
			}
		}()

		// set all joystick's buttons, hats and axis mapping to drone orders
		stick.On(joystick.CirclePress, droneRecording)
		stick.On(joystick.SquarePress, droneTakeOff)
		stick.On(joystick.TrianglePress, droneStop)
		stick.On(joystick.XPress, droneLand)
		stick.On(joystick.LeftX, droneLeftAndRight)
		stick.On(joystick.LeftY, droneForwardAndBackward)
		stick.On(joystick.RightX, droneClockwise)
		stick.On(joystick.RightY, droneUpAndDown)

		gobot.Every(10*time.Millisecond, droneLeftStick)
		gobot.Every(10*time.Millisecond, droneRightStick)
	}

	keyboardWork := func(keys *keyboard.KeyboardDriver) {
		keys.On(keyboard.Key, keyboardQuit)
	}

	robot := gobot.NewRobot("bebop",
		[]gobot.Connection{joystickAdaptor, bebopAdaptor},
		[]gobot.Device{stick, drone},
		droneWork,
	)
	keyRobot := gobot.NewRobot("keyboardbBot",
		[]gobot.Connection{},
		[]gobot.Device{keys},
		keyboardWork,
	)

	Logger.Println("Robot service initiated.")

	gbot.AddRobot(robot)
	Logger.Println("Starting gobot service initiated.")

	Logger.Println("joystickbot service started")
	gbot.AddRobot(keyRobot)

	gbot.Start()

}

type pair struct {
	x float64
	y float64
}

func validatePitch(data float64, offset float64) int {
	value := math.Abs(data) / offset
	if value >= 0.1 {
		if value <= 1.0 {
			return int((float64(int(value*100)) / 100) * 100)
		}
		return 100
	}
	return 0
}

// define the default video output
// use ffmpeg in background, it connects to the drone http stream
func ffmpeg() (stdin io.WriteCloser, stderr io.ReadCloser, err error) {
	Logger.Println("Starting ffmpeg service.")
	ffmpeg := exec.Command("ffmpeg", "-i", "pipe:0", "http://localhost:8090/bebop.ffm")

	stderr, err = ffmpeg.StderrPipe()

	if err != nil {
		return
	}

	stdin, err = ffmpeg.StdinPipe()

	if err != nil {
		return
	}

	if err = ffmpeg.Start(); err != nil {
		return
	}

	go func() {
		for {
			buf, err := ioutil.ReadAll(stderr)
			if err != nil {
				fmt.Println(err)
			}
			if len(buf) > 0 {
				fmt.Println(string(buf))
			}
		}
	}()

	Logger.Println("Ffmpeg service initiated.")
	return stdin, stderr, nil
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

type joystickConfig struct {
	Name    string `json:"name"`
	GUID    string `json:"guid"`
	Axis    []Pair `json:"axis"`
	Buttons []Pair `json:"buttons"`
	Hats    []Hat  `json:"Hats"`
}

func discoverJoystick(gbot *gobot.Gobot, joystickDriver *joystick.JoystickDriver) {
	keys := keyboard.NewKeyboardDriver("keyboard")
	work := func() {
		keys.On(keyboard.Key, keyboardQuit)
	}
	joystickRobot := gobot.NewRobot("joystickbot",
		[]gobot.Connection{},
		[]gobot.Device{joystickDriver},
		func() {

			button := Pair{}

			// define all the joystick configuration
			// robots joystick_driver compliant json format
			joystickConfig := &joystickConfig{}

			// define the circle button
			// take off the drone
			fmt.Println("Click button to take off the drone : ")
			_, id, _, _, _ := sdl_wrapper.SdlEventData()
			button = Pair{Name: utils.RightSuppressButton(joystick.CirclePress), ID: int(id)}
			joystickConfig.Buttons = append(joystickConfig.Buttons, button)

			// define the triangle button
			// stop the drone
			fmt.Println("Click button to stop the drone : ")
			_, id, _, _, _ = sdl_wrapper.SdlEventData()
			button = Pair{Name: utils.RightSuppressButton(joystick.TrianglePress), ID: int(id)}
			joystickConfig.Buttons = append(joystickConfig.Buttons, button)

			// define the square button
			// stop and start the recording video
			fmt.Println("Click button to start and stop the video record of the drone : ")
			_, id, _, _, _ = sdl_wrapper.SdlEventData()
			button = Pair{Name: utils.RightSuppressButton(joystick.SquarePress), ID: int(id)}
			joystickConfig.Buttons = append(joystickConfig.Buttons, button)

			// define the X button
			// landing the drone
			fmt.Println("Click button to land the drone : ")
			_, id, _, _, _ = sdl_wrapper.SdlEventData()
			button = Pair{Name: utils.RightSuppressButton(joystick.XPress), ID: int(id)}
			joystickConfig.Buttons = append(joystickConfig.Buttons, button)

			// define the right stick
			// pilot the drone
			fmt.Println("Click right stick : ")
			_, id, _, _, _ = sdl_wrapper.SdlEventData()
			button = Pair{Name: utils.RightAddStick(joystick.Right), ID: int(id)}
			joystickConfig.Buttons = append(joystickConfig.Buttons, button)

			// define the left stick
			// pilot the drone
			fmt.Println("Click left stick : ")
			_, id, _, _, _ = sdl_wrapper.SdlEventData()
			button = Pair{Name: utils.RightAddStick(joystick.Left), ID: int(id)}
			joystickConfig.Buttons = append(joystickConfig.Buttons, button)

			// missing backward and forward setup
			// missing clockwise and counterclockwise

		})

	keyRobot := gobot.NewRobot("keyboardbBot",
		[]gobot.Connection{},
		[]gobot.Device{keys},
		work,
	)
	gbot.AddRobot(keyRobot)
	Logger.Println("keyboardbot service started")
	gbot.AddRobot(joystickRobot)
	Logger.Println("joystickbot service started")

	gbot.Start()
}

func keyboardQuit(data interface{}) {
	key := data.(keyboard.KeyEvent)
	if key.Key == keyboard.Q {
		fmt.Println("Quiting app")
		Logger.Println("Quiting app")
		os.Exit(0)
	}
}

func droneRecording(_ interface{}) {
	if recording {
		drone.StopRecording()
	} else {
		drone.StartRecording()
	}
	recording = !recording
}

func droneStop(_ interface{}) {
	drone.Stop()
}

func droneUpAndDown(data interface{}) {
	val := float64(data.(int16))
	if rightStick.y != val {
		rightStick.y = val
	}
}

func droneClockwise(data interface{}) {
	val := float64(data.(int16))
	if rightStick.x != val {
		rightStick.x = val
	}
}

func droneForwardAndBackward(data interface{}) {
	val := float64(data.(int16))
	if leftStick.y != val {
		leftStick.y = val
	}
}

func droneLeftAndRight(data interface{}) {
	val := float64(data.(int16))
	if leftStick.x != val {
		leftStick.x = val
	}
}

func droneTakeOff(_ interface{}) {
	drone.HullProtection(true)
	drone.TakeOff()
}

func droneLand(_ interface{}) {
	drone.Land()
}

func droneLeftStick() {
	pair := leftStick
	if pair.y < -10 {
		drone.Forward(validatePitch(pair.y, offset))
	} else if pair.y > 10 {
		drone.Backward(validatePitch(pair.y, offset))
	} else {
		drone.Forward(0)
	}

	if pair.x > 10 {
		drone.Right(validatePitch(pair.x, offset))
	} else if pair.x < -10 {
		drone.Left(validatePitch(pair.x, offset))
	} else {
		drone.Right(0)
	}
}

func droneRightStick() {
	pair := rightStick
	if pair.y < -10 {
		drone.Up(validatePitch(pair.y, offset))
	} else if pair.y > 10 {
		drone.Down(validatePitch(pair.y, offset))
	} else {
		drone.Up(0)
	}

	if pair.x > 20 {
		drone.Clockwise(validatePitch(pair.x, offset))
	} else if pair.x < -20 {
		drone.CounterClockwise(validatePitch(pair.x, offset))
	} else {
		drone.Clockwise(0)
	}
}
