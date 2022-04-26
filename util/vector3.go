package util

import "fmt"

type Vector3 struct {
	Vector [3]float64
}

func NewVector3(x float64, y float64, z float64) Vector3 {
	return Vector3{[3]float64{x, y, z}}
}

func (vec Vector3) ExportASCII() string {
	return fmt.Sprintf("%.6e %.6e %.6e", vec.Vector[X], vec.Vector[Y], vec.Vector[Z])
}

func (vec Vector3) Transform(transform Matrix4) Vector3 {
	vec4 := FromVector3(vec)
	vec4 = transform.MultiplyVector(vec4)
	return vec4.ToVector3()
}

func (vec *Vector3) TransformInPlace(transform Matrix4) {
	result := vec.Transform(transform)
	vec.Vector = result.Vector
}

func (vec Vector3) String() string {
	return fmt.Sprintf("%.6e%s%.6e%s%.6e", vec.Vector[X], delimiterCol, vec.Vector[Y], delimiterCol, vec.Vector[Z])
}

func (vec Vector3) Serialize() string {
	return fmt.Sprintf("%f%s%f%s%f", vec.Vector[X], delimiterCol, vec.Vector[Y], delimiterCol, vec.Vector[Z])
}
