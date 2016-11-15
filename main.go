package main

import (
	"flag"
	"fmt"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/bebop"
	"github.com/hybridgroup/gobot/platforms/joystick"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"time"
)

var APP_VERSION = "1.0"
var testjoystick = flag.String("testjoystick", "", "test your joystick control id")
var Logger *log.Logger

func main() {

	flog, err := os.OpenFile("bebop-ui-control.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		fmt.Fprintf(os.Stdout, "Error while creating log file "+err.Error())
	}
	defer flog.Close()
	Logger = log.New(flog, "Message : ", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Println("Logger service initiated.")

	gbot := gobot.NewGobot()
	gbot.AutoStop = true

	joystickAdaptor := joystick.NewJoystickAdaptor("ps3")
	Logger.Println("Joystick adapter service initiated.")
	stick := joystick.NewJoystickDriver(joystickAdaptor,
		"ps3",
		"./joystick_control_conf.json",
	)

	bebopAdaptor := bebop.NewBebopAdaptor("Drone")
	Logger.Println("Bebop adapter service initiated.")
	drone := bebop.NewBebopDriver(bebopAdaptor, "Drone")
	Logger.Println("Bebop driver service initiated.")

	Logger.Println("Drone and gobots services initiated.")

	work := func() {
		video, _, _ := ffmpeg()

		go func() {
			for {
				if _, err := video.Write(<-drone.Video()); err != nil {
					fmt.Println(err)
					return
				}
			}
		}()

		offset := 32767.0
		rightStick := pair{x: 0, y: 0}
		leftStick := pair{x: 0, y: 0}

		recording := false

		stick.On(joystick.CirclePress, func(data interface{}) {
			if recording {
				drone.StopRecording()
			} else {
				drone.StartRecording()
			}
			recording = !recording
		})

		stick.On(joystick.SquarePress, func(data interface{}) {
			drone.HullProtection(true)
			drone.TakeOff()
		})
		stick.On(joystick.TrianglePress, func(data interface{}) {
			drone.Stop()
		})
		stick.On(joystick.XPress, func(data interface{}) {
			drone.Land()
		})
		stick.On(joystick.LeftX, func(data interface{}) {
			val := float64(data.(int16))
			if leftStick.x != val {
				leftStick.x = val
			}
		})
		stick.On(joystick.LeftY, func(data interface{}) {
			val := float64(data.(int16))
			if leftStick.y != val {
				leftStick.y = val
			}
		})
		stick.On(joystick.RightX, func(data interface{}) {
			val := float64(data.(int16))
			if rightStick.x != val {
				rightStick.x = val
			}
		})
		stick.On(joystick.RightY, func(data interface{}) {
			val := float64(data.(int16))
			if rightStick.y != val {
				rightStick.y = val
			}
		})

		gobot.Every(10*time.Millisecond, func() {
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
		})

		gobot.Every(10*time.Millisecond, func() {
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
		})
	}

	robot := gobot.NewRobot("bebop",
		[]gobot.Connection{joystickAdaptor, bebopAdaptor},
		[]gobot.Device{stick, drone},
		work,
	)
	Logger.Println("Robot service initiated.")
	robot = gbot.AddRobot(robot)

	Logger.Println("Starting gobot service initiated.")

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

func ffmpeg() (stdin io.WriteCloser, stderr io.ReadCloser, err error) {
	Logger.Println("Starting ffmpeg service.")
	ffmpeg := exec.Command("ffmpeg", "-f","dshow","-i", "pipe:0", "http://localhost:8090/bebop.ffm")

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
