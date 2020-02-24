package main

import (
	"github.com/alexflint/go-arg"
	"github.com/robertkrimen/otto"

	"fmt"
	"io/ioutil"
	"log"
	"math"
	"regexp"
	"strconv"
)

type args struct {
	Filename string `arg:"positional,required" help:"JavaScript input file"`
}

func (args) Description() string {
	return ("Compiles go-scad code (JavaScript with a Turtle Graphics-like" +
		" library) into OpenSCAD code.")
}

var vm *otto.Otto

func toJsValue(value interface{}) otto.Value {
	jsValue, err := vm.ToValue(value)
	if err != nil {
		log.Fatal(err)
	}
	return jsValue
}

func toFloat(value otto.Value) float64 {
	if value.IsUndefined() {
		log.Fatal("Undefined value passed to toFloat()")
	}
	floatValue, err := value.ToFloat()
	if err != nil {
		log.Fatal(err)
	}
	return floatValue
}

func toInt(value otto.Value) int {
	if value.IsUndefined() {
		log.Fatal("Undefined value passed to toInt()")
	}
	int64Value, err := value.ToInteger()
	if err != nil {
		log.Fatal(err)
	}
	return int(int64Value)
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func degCos(deg float64) float64 {
	return math.Cos(degToRad(deg))
}

func degSin(deg float64) float64 {
	return math.Sin(degToRad(deg))
}

type TurtlePoint struct {
	X           float64
	Y           float64
	Thickness   float64
	EndCapSides int
}

type TurtlePolygon struct {
	Points   []TurtlePoint
	Headings []float64
}

var stripZeroes *regexp.Regexp

func formatFloat(n float64) string {
	if stripZeroes == nil {
		stripZeroes = regexp.MustCompile(`\.?0+$`)
	}
	str := strconv.FormatFloat(n, 'f', 6, 64)
	str = stripZeroes.ReplaceAllString(str, "")
	if str == "-0" {
		str = "0"
	}
	return str
}

func main() {
	// Parse arguments
	var args args
	arg.MustParse(&args)

	// Read input file
	jsInputBytes, err := ioutil.ReadFile(args.Filename)
	if err != nil {
		log.Fatal(err)
	}

	jsInput := string(jsInputBytes)
	output := jsToScad(jsInput)
	fmt.Print(output)
}

func jsToScad(jsInput string) string {
	output := ""

	outBeginPolygon := func() {
		output += "polygon(points = [\n\t"
	}

	outNewLine := func() {
		output += "\n\t"
	}

	outPoint := func(x float64, y float64, isLast bool) {
		space := " "
		if isLast {
			space = ""
		}
		output += fmt.Sprintf("[%s,%s],%s",
			formatFloat(x),
			formatFloat(y),
			space)
	}

	outEndPolygon := func() {
		output += "\n]);\n"
	}

	// Strip hashbang line if present
	jsInput = regexp.MustCompile(`^#!.*\n`).ReplaceAllString(jsInput, "\n")

	// Set up JavaScript interpreter
	vm = otto.New()

	// Internal state variables
	turtlePendown := false
	turtlePolygons := make([]TurtlePolygon, 0)
	var turtlePenSize float64 = 1
	var turtleEndCapSides int = 60
	var turtleX float64 = 0
	var turtleY float64 = 0
	var turtleHeading float64 = 0
	var turtlePolygon TurtlePolygon

	// Set up functions
	vm.Set("pendown", func(call otto.FunctionCall) otto.Value {
		if !turtlePendown {
			turtlePendown = true
			turtlePolygon = TurtlePolygon{
				Points: []TurtlePoint{{
					X:           turtleX,
					Y:           turtleY,
					Thickness:   turtlePenSize,
					EndCapSides: turtleEndCapSides,
				}},
				Headings: make([]float64, 0),
			}
		}
		return otto.UndefinedValue()
	})
	vm.Set("penup", func(call otto.FunctionCall) otto.Value {
		if turtlePendown {
			turtlePendown = false
			if len(turtlePolygon.Points) != len(turtlePolygon.Headings)+1 {
				log.Fatalf("Bad polygon: points=%d headings=%d",
					len(turtlePolygon.Points),
					len(turtlePolygon.Headings))
			}
			turtlePolygons = append(turtlePolygons, turtlePolygon)
		}
		return otto.UndefinedValue()
	})
	vm.Set("pensize", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsUndefined() {
			return toJsValue(turtlePenSize)
		}
		turtlePenSize = toFloat(call.Argument(0))
		return otto.UndefinedValue()
	})
	vm.Set("end_cap_sides", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsUndefined() {
			return toJsValue(turtleEndCapSides)
		}
		turtleEndCapSides = toInt(call.Argument(0))
		if turtleEndCapSides < 2 || turtleEndCapSides%2 == 1 {
			log.Fatalf("Invalid end_cap_sides value: %d", turtleEndCapSides)
		}
		return otto.UndefinedValue()
	})
	vm.Set("forward", func(call otto.FunctionCall) otto.Value {
		d := toFloat(call.Argument(0))
		turtleX += d * degCos(turtleHeading)
		turtleY += d * degSin(turtleHeading)
		if turtlePendown {
			turtlePolygon.Points = append(turtlePolygon.Points, TurtlePoint{
				X:           turtleX,
				Y:           turtleY,
				Thickness:   turtlePenSize,
				EndCapSides: turtleEndCapSides,
			})
			turtlePolygon.Headings = append(turtlePolygon.Headings, turtleHeading)
		}
		return otto.UndefinedValue()
	})
	vm.Set("right", func(call otto.FunctionCall) otto.Value {
		turtleHeading -= toFloat(call.Argument(0))
		return otto.UndefinedValue()
	})
	vm.Set("left", func(call otto.FunctionCall) otto.Value {
		turtleHeading += toFloat(call.Argument(0))
		return otto.UndefinedValue()
	})

	// Set up aliases
	vm.Run("pd = down = pendown;")
	vm.Run("pu = up = penup;")
	vm.Run("width = pensize;")
	vm.Run("rt = right;")
	vm.Run("lt = left;")

	// Run the script
	_, err := vm.Run(jsInput)
	if err != nil {
		if jsErr, ok := err.(*otto.Error); ok {
			log.Fatalf("JavaScript error: %s", jsErr.String())
		} else {
			log.Fatal("JavaScript error: ", err)
		}
	}

	// Turn the results into OpenSCAD output
	for _, polygon := range turtlePolygons {
		outBeginPolygon()

		if len(polygon.Points) == 1 {
			// Degenerate case: just draw an end cap
			point := polygon.Points[0]
			for j := 0; j < point.EndCapSides; j++ {
				angle := float64(j) * 360 / float64(point.EndCapSides)
				outPoint(
					point.X+point.Thickness/2*degCos(angle),
					point.Y+point.Thickness/2*degSin(angle),
					j == point.EndCapSides-1)
			}
			outEndPolygon()
			continue
		}

		// Loop around the polygon's coordinates twice (first in ascending
		// order, then in descending order) to draw the "left" (d == 1) and
		// "right" (d == -1) edges of its pen strokes, in a clockwise fashion.
		d := 1
		i := 0
		for {
			point := polygon.Points[i]
			if i == 0 {
				// Draw begin cap
				headingBegin := polygon.Headings[0]
				for j := 0; j <= point.EndCapSides/2; j++ {
					angle := headingBegin - 90 - float64(j)*360/float64(point.EndCapSides)
					outPoint(
						point.X+point.Thickness/2*degCos(angle),
						point.Y+point.Thickness/2*degSin(angle),
						j == point.EndCapSides/2)
				}
				outNewLine()
			} else if i == len(polygon.Points)-1 {
				// Draw end cap
				if len(polygon.Points) > 2 {
					outNewLine()
				}
				headingEnd := polygon.Headings[i-1]
				for j := 0; j <= point.EndCapSides/2; j++ {
					angle := headingEnd + 90 - float64(j)*360/float64(point.EndCapSides)
					outPoint(
						point.X+point.Thickness/2*degCos(angle),
						point.Y+point.Thickness/2*degSin(angle),
						j == point.EndCapSides/2)
				}
				if len(polygon.Points) > 2 {
					outNewLine()
				}
			} else {
				// Draw left or right side
				heading := (point.Heading+headingPrev)/2 + 90
				heading = heading * math.Pi / 180
				outPoint(
					point.X+float64(d)*point.Thickness/2*math.Cos(heading),
					point.Y+float64(d)*point.Thickness/2*math.Sin(heading),
					((i == len(polygon)-2 && d == 1) || (i == 1 && d == -1)))
			}

			if i == len(polygon.Points)-1 && d == 1 {
				d = -1
			}
			if i == 1 && d == -1 {
				break
			} else {
				i += d
			}
		}

		outEndPolygon()
	}

	return output
}
