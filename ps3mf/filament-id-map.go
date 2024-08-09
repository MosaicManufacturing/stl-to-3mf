package ps3mf

import "encoding/json"

// FilamentIDMap contains a mapping from project input index to Element filamentId
// (only used for Element processing)
type FilamentIDMap map[byte]byte

type filamentIDJsonData struct {
	IDs [][]byte `json:"filamentIds"`
}

func UnmarshalFilamentIds(serialized string) (FilamentIDMap, error) {
	var parsed filamentIDJsonData
	if err := json.Unmarshal([]byte(serialized), &parsed); err != nil {
		return nil, err
	}

	ids := make(FilamentIDMap)
	for _, pair := range parsed.IDs {
		key := pair[0]
		value := pair[1]
		ids[key] = value
	}

	return ids, nil
}
