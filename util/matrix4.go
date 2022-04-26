package util

import (
	"fmt"
	"github.com/hpinc/go3mf"
	"math"
	"strconv"
	"strings"
)

const delimiterRow = "|"
const delimiterCol = ","

const (
	X = 0
	Y = 1
	Z = 2
	W = 3
)

type Matrix4 struct {
	Matrix [4]Vector4
}

func NewMatrix4() Matrix4 {
	return Matrix4{
		Matrix: [4]Vector4{
			NewVector4(0, 0, 0, 0),
			NewVector4(0, 0, 0, 0),
			NewVector4(0, 0, 0, 0),
			NewVector4(0, 0, 0, 0),
		},
	}
}

func M4Identity() Matrix4 {
	m := NewMatrix4()
	m.Matrix[0].Vector[X] = 1
	m.Matrix[1].Vector[Y] = 1
	m.Matrix[2].Vector[Z] = 1
	m.Matrix[3].Vector[W] = 1
	return m
}

func M4Translate(dx float64, dy float64, dz float64) Matrix4 {
	m := M4Identity()
	m.Matrix[3].Vector[X] = dx
	m.Matrix[3].Vector[Y] = dy
	m.Matrix[3].Vector[Z] = dz
	return m
}

func M4Scale(sx float64, sy float64, sz float64) Matrix4 {
	m := M4Identity()
	m.Matrix[0].Vector[X] = sx
	m.Matrix[1].Vector[Y] = sy
	m.Matrix[2].Vector[Z] = sz
	return m
}

func M4RotateX(theta float64) Matrix4 {
	m := M4Identity()
	m.Matrix[1].Vector[Y] = math.Cos(theta)
	m.Matrix[1].Vector[Z] = math.Sin(theta)
	m.Matrix[2].Vector[Y] = -math.Sin(theta)
	m.Matrix[2].Vector[Z] = math.Cos(theta)
	return m
}

func M4RotateY(theta float64) Matrix4 {
	m := M4Identity()
	m.Matrix[0].Vector[X] = math.Cos(theta)
	m.Matrix[0].Vector[Z] = -math.Sin(theta)
	m.Matrix[2].Vector[X] = math.Sin(theta)
	m.Matrix[2].Vector[Z] = math.Cos(theta)
	return m
}

func M4RotateZ(theta float64) Matrix4 {
	m := M4Identity()
	m.Matrix[0].Vector[X] = math.Cos(theta)
	m.Matrix[0].Vector[Y] = math.Sin(theta)
	m.Matrix[1].Vector[X] = -math.Sin(theta)
	m.Matrix[1].Vector[Y] = math.Cos(theta)
	return m
}

func (m Matrix4) Multiply(matrix Matrix4) Matrix4 {
	a := m.Matrix
	b := matrix.Matrix
	result := NewMatrix4()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			product := float64(0)
			for k := 0; k < 4; k++ {
				product += a[k].Vector[j] * b[i].Vector[k]
			}
			result.Matrix[i].Vector[j] = product
		}
	}
	return result
}

func (m Matrix4) MultiplyVector(v Vector4) Vector4 {
	a := m.Matrix
	b := v.Vector
	result := NewVector4(0, 0, 0, 0)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result.Vector[i] += a[j].Vector[i] * b[j]
		}
	}
	return result
}

func (m Matrix4) Serialize() string {
	var output []string
	for i := 0; i < 4; i++ {
		var line []string
		for j := 0; j < 4; j++ {
			line = append(line, fmt.Sprintf("%f", m.Matrix[i].Vector[j]))
		}
		output = append(output, strings.Join(line, delimiterCol))
	}
	return strings.Join(output, delimiterRow)
}

func UnserializeMatrix4(str string) (Matrix4, error) {
	m := NewMatrix4()
	lines := strings.Split(str, delimiterRow)

	if len(lines) != 4 {
		return m, fmt.Errorf("expected 4 rows in serialized Matrix4, found %d", len(lines))
	}
	for i := 0; i < 4; i++ {
		line := strings.Split(lines[i], delimiterCol)
		if len(line) != 4 {
			return m, fmt.Errorf("expected 4 columns in serialized Matrix4, found %d", len(line))
		}
		for j := 0; j < 4; j++ {
			value, err := strconv.ParseFloat(line[j], 64)
			if err != nil {
				return m, err
			}
			m.Matrix[i].Vector[j] = value
		}
	}
	return m, nil
}

func (m Matrix4) String() string {
	str := ""
	for i := 0; i < 4; i++ {
		str += "[ "
		for j := 0; j < 3; j++ {
			str += fmt.Sprintf("%f", m.Matrix[i].Vector[j]) + ", "
		}
		str += fmt.Sprintf("%f", m.Matrix[i].Vector[3])
		str += " ]\n"
	}
	return str
}

func (m Matrix4) To3MF() go3mf.Matrix {
	return go3mf.Matrix{
		float32(m.Matrix[0].Vector[X]), float32(m.Matrix[0].Vector[Y]), float32(m.Matrix[0].Vector[Z]), float32(m.Matrix[0].Vector[W]),
		float32(m.Matrix[1].Vector[X]), float32(m.Matrix[1].Vector[Y]), float32(m.Matrix[1].Vector[Z]), float32(m.Matrix[1].Vector[W]),
		float32(m.Matrix[2].Vector[X]), float32(m.Matrix[2].Vector[Y]), float32(m.Matrix[2].Vector[Z]), float32(m.Matrix[2].Vector[W]),
		float32(m.Matrix[3].Vector[X]), float32(m.Matrix[3].Vector[Y]), float32(m.Matrix[3].Vector[Z]), float32(m.Matrix[3].Vector[W]),
	}
}
