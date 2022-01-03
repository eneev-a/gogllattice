// Copyright 2022 Alan Eneev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Renders a textured spinning cube using GLFW 3 and OpenGL 4.1 core forward-compatible profile.
package main

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	windowWidth  = 800
	windowHeight = 600
)

var (
	x    = mgl32.Vec3{1, 0, 0}
	y    = mgl32.Vec3{0, 1, 0}
	z    = mgl32.Vec3{0, 0, 1}
	zero = mgl32.Vec3{}
)

type FrameTimer struct {
	prevTime   float64
	elapsed    float64
	checkPoint float64
	frames     int32
	mspf       float32
}

func (ft *FrameTimer) OnFrame() {
	if ft.prevTime == 0 {
		ft.prevTime = glfw.GetTime()
		return
	}

	period := 1.0
	time := glfw.GetTime()
	ft.elapsed = time - ft.prevTime
	ft.prevTime = time
	if time >= ft.checkPoint {
		dt := (time - (ft.checkPoint - period))
		ft.mspf = 1000 * float32(dt) / float32(ft.frames)
		ft.checkPoint = time + period
		ft.frames = 0
	}
	ft.frames++
}

type State struct {
	camSpeed      mgl32.Vec3
	camPos        mgl32.Vec3
	rotationSpeed mgl32.Vec3
	cameraUniform int32
	shiftUniform  int32
	camEnabled    bool

	prevCursorX, prevCursorY float64
	dx, dy                   float64

	roll  float32
	pitch float32
	yaw   float32

	frameTimer FrameTimer

	w *glfw.Window

	count int
}

func NewState(w *glfw.Window) *State {
	return &State{
		camPos: mgl32.Vec3{-41.5, -43.5, -37.5},
		pitch:  mgl32.DegToRad(21.5),
		yaw:    mgl32.DegToRad(-135),
		w:      w,
	}
}

func (s *State) Update(w *glfw.Window) {
	s.frameTimer.OnFrame()
	dt := s.frameTimer.elapsed
	if dt == 0 {
		return
	}

	sensitivity := float32(0.001)

	s.roll = 0
	s.pitch = normAngle(s.pitch + float32(-s.dy)*sensitivity)
	s.pitch = mgl32.Clamp(s.pitch, -math.Pi/2, math.Pi/2)
	s.yaw = normAngle(s.yaw + float32(-s.dx)*sensitivity)
	s.dx, s.dy = 0, 0

	q := mgl32.AnglesToQuat(s.roll, s.yaw, s.pitch, mgl32.ZYX)
	s.camPos = s.camPos.Add(q.Rotate(s.camSpeed).Mul(float32(dt)))

	camera := mgl32.Ident4()
	camera = q.Mat4().Mul4(camera)
	camera = mgl32.Translate3D(s.camPos[0], s.camPos[1], s.camPos[2]).Mul4(camera)
	camera = camera.Inv()

	gl.UniformMatrix4fv(s.cameraUniform, 1, false, &camera[0])

	gl.Uniform1f(s.shiftUniform, float32(1+math.Sin(s.frameTimer.prevTime/2))/2/4+0.002)
}

func (s *State) OnKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press && action != glfw.Release {
		return
	}

	camSpeed := float32(5.0)
	if (mods & glfw.ModControl) > 0 {
		camSpeed = 20
	}
	if (mods & glfw.ModShift) > 0 {
		camSpeed = 0.1
	}
	mul := float32(1.0)
	if action == glfw.Release {
		mul = 0
	}

	rotStep := float32(math.Pi / 16)

	switch key {

	case glfw.KeyA:
		s.camSpeed[0] = -camSpeed * mul
	case glfw.KeyD:
		s.camSpeed[0] = +camSpeed * mul
	case glfw.KeyW:
		s.camSpeed[2] = -camSpeed * mul
	case glfw.KeyS:
		s.camSpeed[2] = +camSpeed * mul
	case glfw.KeySpace:
		s.camSpeed[1] = +camSpeed * mul
	case glfw.KeyZ:
		s.camSpeed[1] = -camSpeed * mul
	case glfw.KeyUp:
		s.pitch += mul * rotStep
	case glfw.KeyDown:
		s.pitch -= mul * rotStep
	case glfw.KeyLeft:
		s.yaw += mul * rotStep
	case glfw.KeyRight:
		s.yaw -= mul * rotStep

	case glfw.KeyC:
		s.roll = 0
		s.pitch = mgl32.DegToRad(-34.5)
		s.yaw = mgl32.DegToRad(45)
		s.camPos = mgl32.Vec3{30, 30, 30}
	case glfw.KeyEscape:
		log.Fatal("ESC pressed")
	}
}

func (s *State) OnCursorEnter(w *glfw.Window, entered bool) {
	s.camEnabled = entered
	if entered {
		s.prevCursorX, s.prevCursorY = w.GetCursorPos()
	}
}

func (s *State) OnCursorPos(w *glfw.Window, xpos, ypos float64) {
	if !s.camEnabled {
		return
	}
	s.dx += (xpos - s.prevCursorX)
	s.dy += (ypos - s.prevCursorY)
	s.prevCursorX = xpos
	s.prevCursorY = ypos
}

func (s *State) RenderToTerm() {

	fmt.Printf("ms per frame: %v\n", s.frameTimer.mspf)

	fmt.Println("Camera:")
	fmt.Printf("  roll: %v (%v)\n", s.roll, mgl32.RadToDeg(s.roll))
	fmt.Printf("  pitch: %v (%v)\n", s.pitch, mgl32.RadToDeg(s.pitch))
	fmt.Printf("  yaw: %v (%v)\n", s.yaw, mgl32.RadToDeg(s.yaw))
	fmt.Printf("  x: %v\n", s.camPos[0])
	fmt.Printf("  y: %v\n", s.camPos[1])
	fmt.Printf("  z: %v\n", s.camPos[2])

	fmt.Println("Mouse:")
	fmt.Printf("  x: %v\n", s.prevCursorX)
	fmt.Printf("  y: %v\n", s.prevCursorY)
	fmt.Println("Triangle count:", s.count)
	fmt.Println("Time:", s.frameTimer.prevTime)
}

func normAngle(rad float32) float32 {
	for rad > math.Pi {
		rad -= 2 * math.Pi
	}
	for rad < -math.Pi {
		rad += 2 * math.Pi
	}
	return rad
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func makeVerts(t float64) []float32 {
	d := 30
	dd := 1 / float32(2*d+1)

	t = t / 20

	verts := make([]float32, (d+1)*(d+1)*(d+1)*9*3*12)
	for x := -d; x <= d; x++ {
		for y := -d; y <= d; y++ {
			for z := -d; z <= d; z++ {

				r := dd * float32(x+d)
				g := dd * float32(y+d)
				b := dd * float32(z+d)
				x, y, z := float32(x), float32(y), float32(z)
				const w = 1
				verts = append(verts, []float32{
					// Top
					x - w/2, y + w/2, z - w/2, r, g, b, 1, -1, 1,
					x + w/2, y + w/2, z + w/2, r, g, b, -1, -1, -1,
					x + w/2, y + w/2, z - w/2, r, g, b, -1, -1, 1,
					x - w/2, y + w/2, z - w/2, r, g, b, 1, -1, 1,
					x + w/2, y + w/2, z + w/2, r, g, b, -1, -1, -1,
					x - w/2, y + w/2, z + w/2, r, g, b, 1, -1, -1,

					// Bottom
					x - w/2, y - w/2, z - w/2, r, g, b, 1, 1, 1,
					x + w/2, y - w/2, z + w/2, r, g, b, -1, 1, -1,
					x + w/2, y - w/2, z - w/2, r, g, b, -1, 1, 1,
					x - w/2, y - w/2, z - w/2, r, g, b, 1, 1, 1,
					x + w/2, y - w/2, z + w/2, r, g, b, -1, 1, -1,
					x - w/2, y - w/2, z + w/2, r, g, b, 1, 1, -1,

					// Front
					x - w/2, y + w/2, z + w/2, r, g, b, 1, -1, -1,
					x + w/2, y + w/2, z + w/2, r, g, b, -1, -1, -1,
					x + w/2, y - w/2, z + w/2, r, g, b, -1, 1, -1,
					x - w/2, y + w/2, z + w/2, r, g, b, 1, -1, -1,
					x - w/2, y - w/2, z + w/2, r, g, b, 1, 1, -1,
					x + w/2, y - w/2, z + w/2, r, g, b, -1, 1, -1,

					// Back
					x - w/2, y + w/2, z - w/2, r, g, b, 1, -1, 1,
					x + w/2, y + w/2, z - w/2, r, g, b, -1, -1, 1,
					x + w/2, y - w/2, z - w/2, r, g, b, -1, 1, 1,
					x - w/2, y + w/2, z - w/2, r, g, b, 1, -1, 1,
					x - w/2, y - w/2, z - w/2, r, g, b, 1, 1, 1,
					x + w/2, y - w/2, z - w/2, r, g, b, -1, 1, 1,

					// Left
					x - w/2, y + w/2, z - w/2, r, g, b, 1, -1, 1,
					x - w/2, y + w/2, z + w/2, r, g, b, 1, -1, -1,
					x - w/2, y - w/2, z + w/2, r, g, b, 1, 1, -1,
					x - w/2, y + w/2, z - w/2, r, g, b, 1, -1, 1,
					x - w/2, y - w/2, z + w/2, r, g, b, 1, 1, -1,
					x - w/2, y - w/2, z - w/2, r, g, b, 1, 1, 1,

					// Right
					x + w/2, y + w/2, z - w/2, r, g, b, -1, -1, 1,
					x + w/2, y + w/2, z + w/2, r, g, b, -1, -1, -1,
					x + w/2, y - w/2, z + w/2, r, g, b, -1, 1, -1,
					x + w/2, y + w/2, z - w/2, r, g, b, -1, -1, 1,
					x + w/2, y - w/2, z + w/2, r, g, b, -1, 1, -1,
					x + w/2, y - w/2, z - w/2, r, g, b, -1, 1, 1,
				}...)
			}
		}
	}
	return verts
}

func main() {

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Samples, 8)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	m := glfw.GetPrimaryMonitor()
	vm := m.GetVideoMode()
	window, err := glfw.CreateWindow(vm.Width, vm.Height, "Render", nil, nil)
	window.SetMonitor(glfw.GetPrimaryMonitor(), 0, 0, vm.Width, vm.Height, vm.RefreshRate)
	s := NewState(window)
	go func() {
		for {
			s.RenderToTerm()
			time.Sleep(time.Duration(1000) * time.Millisecond)
		}
	}()

	window.SetKeyCallback(s.OnKey)
	window.SetCursorEnterCallback(s.OnCursorEnter)
	window.SetCursorPosCallback(s.OnCursorPos)
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	if glfw.RawMouseMotionSupported() {
		window.SetInputMode(glfw.RawMouseMotion, glfw.True)
	}

	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	w, h := window.GetSize()
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(w)/float32(h), 0.01, 500.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	shiftUniform := gl.GetUniformLocation(program, gl.Str("shift\x00"))
	gl.Uniform1f(shiftUniform, 1)

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	verts := makeVerts(s.frameTimer.prevTime)
	s.count = len(verts) / 3 / 3
	gl.BufferData(gl.ARRAY_BUFFER, len(verts)*4, gl.Ptr(verts), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointerWithOffset(vertAttrib, 3, gl.FLOAT, false, 9*4, 0)

	colorAttrib := uint32(gl.GetAttribLocation(program, gl.Str("color\x00")))
	gl.EnableVertexAttribArray(colorAttrib)
	gl.VertexAttribPointerWithOffset(colorAttrib, 3, gl.FLOAT, false, 9*4, 3*4)

	shiftDirAttrib := uint32(gl.GetAttribLocation(program, gl.Str("shiftDir\x00")))
	gl.EnableVertexAttribArray(shiftDirAttrib)
	gl.VertexAttribPointerWithOffset(shiftDirAttrib, 3, gl.FLOAT, false, 9*4, 6*4)

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	s.cameraUniform = cameraUniform
	s.shiftUniform = shiftUniform

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		s.Update(window)

		// Render
		gl.UseProgram(program)

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(verts)/9))

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

var vertexShader = `
#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float shift;

in vec3 vert;
in vec3 color;
in vec3 shiftDir;
out vec3 fragColor;

void main() {
    gl_Position = projection * camera * model * vec4(shiftDir * shift + vert, 1);
		fragColor = color;
}
` + "\x00"

var fragmentShader = `
#version 330

in vec3 fragColor;
out vec4 outputColor;

void main() {
    outputColor = vec4(fragColor.xyz, 0);
}
` + "\x00"
