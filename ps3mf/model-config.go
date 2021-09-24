package ps3mf

import (
	"encoding/xml"
	"strconv"
)

type ModelConfig struct {
	XMLName xml.Name `xml:"config"`
	Objects []ModelConfigObject `xml:"object"`
}

type ModelConfigObject struct {
	XMLName xml.Name `xml:"object"`
	Id string `xml:"id,attr"`
	InstancesCount string `xml:"instances_count,attr"`
	Metadata []ModelConfigMeta `xml:"metadata"`
	Volume ModelConfigVolume `xml:"volume"`
}

type ModelConfigVolume struct {
	XMLName xml.Name `xml:"volume"`
	FirstId string `xml:"firstid,attr"`
	LastId string `xml:"lastid,attr"`
	Metadata []ModelConfigMeta `xml:"metadata"`
}

type ModelConfigMeta struct {
	XMLName xml.Name `xml:"metadata"`
	Type string `xml:"type,attr"`
	Key string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

func GetModelConfigMeta(typ, key, value string) ModelConfigMeta {
	return ModelConfigMeta{
		Type: typ,
		Key: key,
		Value: value,
	}
}

func boolToIntString(b bool) string {
	if b { return "1" }
	return "0"
}

func (m *Bundle) GetModelConfig() ModelConfig {
	// TODO: add future support for infill transitioning (purge_to_infill)
	//  and model transitioning (purge_to_models)
	config := ModelConfig{
		Objects: make([]ModelConfigObject, 0, len(m.Model.Resources.Objects)),
	}
	for idx := range m.Model.Resources.Objects {
		id := strconv.Itoa(int(m.Model.Resources.Objects[idx].ID))
		objectConfig := ModelConfigObject{
			Id: id,
			InstancesCount: "1",
			Metadata: []ModelConfigMeta{
				GetModelConfigMeta("object", "name", "model"),
				GetModelConfigMeta("object", "extruder", m.Extruders[idx]),
				GetModelConfigMeta("object", "wipe_into_infill", boolToIntString(m.WipeIntoInfill[idx])),
				GetModelConfigMeta("object", "wipe_into_objects", boolToIntString(m.WipeIntoModel[idx])),
			},
			Volume: ModelConfigVolume{
				XMLName:  xml.Name{},
				FirstId:  "0",
				LastId:   strconv.Itoa(len(m.Model.Resources.Objects[idx].Mesh.Triangles) - 1),
				Metadata: make([]ModelConfigMeta, 0),
			},
		}
		config.Objects = append(config.Objects, objectConfig)
	}
	return config
}
