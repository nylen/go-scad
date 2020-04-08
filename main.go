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
	"strings"
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

func toString(value otto.Value) string {
	if value.IsUndefined() {
		log.Fatal("Undefined value passed to toString()")
	}
	stringValue, err := value.ToString()
	if err != nil {
		log.Fatal(err)
	}
	return stringValue
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
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

	indentLevel := 0

	outBeginPolygon := func() {
		output += strings.Repeat("\t", indentLevel) +
			"polygon(points = [\n" +
			strings.Repeat("\t", indentLevel+1)
	}

	outNewLine := func() {
		output += "\n" + strings.Repeat("\t", indentLevel+1)
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
		output += "\n" + strings.Repeat("\t", indentLevel) + "]);\n"
	}

	outBeginBlock := func(wrapper string) {
		output += strings.Repeat("\t", indentLevel) + wrapper + " {\n"
		indentLevel += 1
	}

	outEndBlock := func() {
		indentLevel -= 1
		output += strings.Repeat("\t", indentLevel) + "}\n"
	}

	writePolygon := func(polygon TurtlePolygon) {
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
			return
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
				// Join together two pen strokes
				var headingPrev float64
				var headingNext float64
				if d == 1 {
					headingPrev = polygon.Headings[i-1]
					headingNext = polygon.Headings[i]
				} else {
					headingPrev = polygon.Headings[i]
					headingNext = polygon.Headings[i-1]
				}
				isLastPoint :=
					((i == len(polygon.Points)-2 && d == 1) || (i == 1 && d == -1))
				if headingPrev == headingNext {
					// Degenerate case: both segments being joined have the same
					// heading.  The end of the current pen-stroke is the start
					// of the next pen-stroke, no need to calculate more.
					heading := headingPrev + float64(90*d)
					outPoint(
						point.X+point.Thickness/2*degCos(heading),
						point.Y+point.Thickness/2*degSin(heading),
						isLastPoint)
				} else {
					// Need to calculate the point marked with an 'x' in the
					// diagram below, which is the intersection of the edges of
					// the current pen-stroke (line between points 1-2) and the
					// next pen-stroke (line between points 3-4):
					//
					//       / .  4
					//   ----    /
					//   .   .  /
					//  1------x2
					//        3
					//
					pointPrev := polygon.Points[i-d]
					pointNext := polygon.Points[i+d]
					headingEdgePrev := headingPrev + float64(90*d)
					headingEdgeNext := headingNext + float64(90*d)
					// Point 1
					x1 := pointPrev.X + pointPrev.Thickness/2*degCos(headingEdgePrev)
					y1 := pointPrev.Y + pointPrev.Thickness/2*degSin(headingEdgePrev)
					// Point 2
					x2 := point.X + point.Thickness/2*degCos(headingEdgePrev)
					y2 := point.Y + point.Thickness/2*degSin(headingEdgePrev)
					// Point 3
					x3 := point.X + point.Thickness/2*degCos(headingEdgeNext)
					y3 := point.Y + point.Thickness/2*degSin(headingEdgeNext)
					// Point 4
					x4 := pointNext.X + pointNext.Thickness/2*degCos(headingEdgeNext)
					y4 := pointNext.Y + pointNext.Thickness/2*degSin(headingEdgeNext)
					// Calculation
					// https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_points_on_each_line
					denom := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
					x := ((x1*y2-y1*x2)*(x3-x4) - (x1-x2)*(x3*y4-y3*x4)) / denom
					y := ((x1*y2-y1*x2)*(y3-y4) - (y1-y2)*(x3*y4-y3*x4)) / denom
					outPoint(x, y, isLastPoint)
				}
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

	// Strip hashbang line if present
	jsInput = regexp.MustCompile(`^#!.*\n`).ReplaceAllString(jsInput, "\n")

	// Set up JavaScript interpreter
	vm = otto.New()

	// Internal state variables
	turtlePendown := false
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
			writePolygon(turtlePolygon)
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
	vm.Set("setpos", func(call otto.FunctionCall) otto.Value {
		x := toFloat(call.Argument(0))
		y := toFloat(call.Argument(1))
		thisHeading := radToDeg(math.Atan2(y-turtleY, x-turtleX))
		turtleX = x
		turtleY = y
		if turtlePendown {
			turtlePolygon.Points = append(turtlePolygon.Points, TurtlePoint{
				X:           turtleX,
				Y:           turtleY,
				Thickness:   turtlePenSize,
				EndCapSides: turtleEndCapSides,
			})
			turtlePolygon.Headings = append(turtlePolygon.Headings, thisHeading)
		}
		return otto.UndefinedValue()
	})
	vm.Set("heading", func(call otto.FunctionCall) otto.Value {
		return toJsValue(turtleHeading)
	})
	vm.Set("wrap", func(call otto.FunctionCall) otto.Value {
		outBeginBlock(toString(call.Argument(0)))
		call.Argument(1).Call(otto.UndefinedValue())
		outEndBlock()
		return otto.UndefinedValue()
	})

	// Set up aliases
	vm.Run("pd = down = pendown;")
	vm.Run("pu = up = penup;")
	vm.Run("width = pensize;")
	vm.Run("rt = right;")
	vm.Run("lt = left;")
	vm.Run("setposition = setpos;") // Note, no `goto` alias (reserved word)

	// Run the script
	_, err := vm.Run(jsInput)
	if err != nil {
		if jsErr, ok := err.(*otto.Error); ok {
			log.Fatalf("JavaScript error: %s", jsErr.String())
		} else {
			log.Fatal("JavaScript error: ", err)
		}
	}


	return output
}
