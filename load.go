package main

//func addColorToModel(model *go3mf.Model, color util.RLE) {
//	currentRunIndex := -1
//	currentRunLength := 0
//	currentColor := 0
//
//	for triIdx := range model.Resources.Objects[0].Mesh.Triangles {
//		if currentRunLength <= 0 {
//			if currentRunIndex < len(color) {
//				currentRunIndex++
//				currentRunLength = int(color[currentRunIndex].Length)
//				currentColor = int(color[currentRunIndex].Value) + 1
//			}
//		}
//		model.Resources.Objects[0].Mesh.Triangles[triIdx].AnyAttr = currentColor
//	}
//
//		attributes = uint16(currentColor)
//		currentRunLength--
//
//}

//func addSupportsToModel(model *go3mf.Model, supports util.RLE) {
//}
