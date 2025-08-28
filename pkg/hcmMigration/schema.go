package hcmmigration

type Profiles struct {
	ProfileList ProfileList `xml:"ProfileList"`
}

type ProfileList struct {
	Profiles []Profile `xml:"Profile"`
}

type Profile struct {
	Name        string `xml:"Name"`
	Attempts    int    `xml:"Attempts"`
	ActiveSplit int    `xml:"ActiveSplit"`
	Rows        Rows   `xml:"Rows"`
}

type Rows struct {
	ProfileRow []ProfileRow `xml:"ProfileRow"`
}

type ProfileRow struct {
	Title        string `xml:"Title"`
	Hits         int    `xml:"Hits"`
	WayHits      int    `xml:"WayHits"`
	PB           int    `xml:"PB"`
	Duration     int    `xml:"Duration"`
	DurationPB   int    `xml:"DurationPB"`
	DurationGold int    `xml:"DurationGold"`
}
