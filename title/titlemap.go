package title

import "strings"

type Map struct {
	//XMLName xml.Name `xml:"map"`
	Version      string        `xml:"version,attr"`
	Width        int           `xml:"width,attr"`
	Height       int           `xml:"height,attr"`
	ObjectGroups []ObjectGroup `xml:"objectgroup"`
}

func (m *Map) Init() error {
	for _, v := range m.ObjectGroups {
		switch strings.ToLower(v.Name) {
		case "walkable":
			if err := v.InitWalkable(); err != nil {
				return err
			}
		}
	}
	return nil
}
