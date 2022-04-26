package util

type Vector4 struct {
	Vector [4]float64
}

func NewVector4(x float64, y float64, z float64, w float64) Vector4 {
	return Vector4{[4]float64{x, y, z, w}}
}

func FromVector3(vec Vector3) Vector4 {
	return NewVector4(vec.Vector[X], vec.Vector[Y], vec.Vector[Z], 1)
}

func (vec Vector4) ToVector3() Vector3 {
	return NewVector3(vec.Vector[X], vec.Vector[Y], vec.Vector[Z])
}
