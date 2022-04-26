package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type BoundingBox struct {
	Min Vector3
	Max Vector3
}

func NewBoundingBox() BoundingBox {
	return BoundingBox{
		Min: NewVector3(math.Inf(1), math.Inf(1), math.Inf(1)),
		Max: NewVector3(math.Inf(-1), math.Inf(-1), math.Inf(-1)),
	}
}

func (bbox *BoundingBox) ExpandByPoint(point Vector3) {
	bbox.Min.Vector[X] = math.Min(bbox.Min.Vector[X], point.Vector[X])
	bbox.Min.Vector[Y] = math.Min(bbox.Min.Vector[Y], point.Vector[Y])
	bbox.Min.Vector[Z] = math.Min(bbox.Min.Vector[Z], point.Vector[Z])
	bbox.Max.Vector[X] = math.Max(bbox.Max.Vector[X], point.Vector[X])
	bbox.Max.Vector[Y] = math.Max(bbox.Max.Vector[Y], point.Vector[Y])
	bbox.Max.Vector[Z] = math.Max(bbox.Max.Vector[Z], point.Vector[Z])
}

func (bbox *BoundingBox) ExpandByBox(box BoundingBox) {
	bbox.Min.Vector[X] = math.Min(bbox.Min.Vector[X], box.Min.Vector[X])
	bbox.Min.Vector[Y] = math.Min(bbox.Min.Vector[Y], box.Min.Vector[Y])
	bbox.Min.Vector[Z] = math.Min(bbox.Min.Vector[Z], box.Min.Vector[Z])
	bbox.Max.Vector[X] = math.Max(bbox.Max.Vector[X], box.Max.Vector[X])
	bbox.Max.Vector[Y] = math.Max(bbox.Max.Vector[Y], box.Max.Vector[Y])
	bbox.Max.Vector[Z] = math.Max(bbox.Max.Vector[Z], box.Max.Vector[Z])
}

func (bbox BoundingBox) GetCenter() Vector3 {
	return NewVector3(
		(bbox.Min.Vector[X]+bbox.Max.Vector[X])/2,
		(bbox.Min.Vector[Y]+bbox.Max.Vector[Y])/2,
		(bbox.Min.Vector[Z]+bbox.Max.Vector[Z])/2,
	)
}

func (bbox BoundingBox) Serialize() string {
	serializedMin := bbox.Min.Serialize()
	serializedMax := bbox.Max.Serialize()
	return fmt.Sprint(serializedMin, delimiterRow, serializedMax)
}

func UnserializeBoundingBox(str string) (BoundingBox, error) {
	bbox := NewBoundingBox()
	lines := strings.Split(str, delimiterRow)
	if len(lines) != 2 {
		return bbox, fmt.Errorf("expected 2 rows in serialized BoundingBox, found %d", len(lines))
	}
	for i := 0; i < 2; i++ {
		line := strings.Split(lines[i], delimiterCol)
		if len(line) != 3 {
			return bbox, fmt.Errorf("expected 3 columns in serialized BoundingBox, found %d", len(line))
		}
		for j := 0; j < 3; j++ {
			value, err := strconv.ParseFloat(line[j], 64)
			if err != nil {
				return bbox, err
			}
			if i == 0 {
				bbox.Min.Vector[j] = value
			} else {
				bbox.Max.Vector[j] = value
			}
		}
	}
	return bbox, nil
}
