package main

import (
	"fmt"
	"strings"
)

/*
<?xml version="1.0" encoding="UTF-8"?>
<kml>
	<Document>
		<name>Great Himalaya Trail</name>
        <description>...</description>
        <visibility>1</visibility>
        <open>1</open>
        <Style id="route_red">
            <LineStyle>
            <color>961400FF</color>
            <width>4</width>
            </LineStyle>
        </Style>
        ...

        <Folder>
            <name>Waypoints</name>
            <description>...</description>
            <visibility>1</visibility>
            <open>0</open>

            <Placemark>
                <name>...</name>
                <visibility>1</visibility>
                <open>0</open>
                <description>...</description>
                <Point>
                    <coordinates>
                        lat,lon,ele
                    </coordinates>
                </Point>
            </Placemark>
            ...
		</Folder>

		<Folder>
            <name>Routes</name>
            <description>...</description>
            <visibility>1</visibility>
            <open>0</open>

			<Placemark>
                <visibility>0</visibility>
                <open>0</open>
                <styleUrl>#route_red</styleUrl>
                <name>...</name>
                <description>...</description>
                <LineString>
                    <extrude>true</extrude>
                    <tessellate>true</tessellate>
                    <altitudeMode>clampToGround</altitudeMode>
                    <coordinates>
                        lat,lon,ele lat,lon,ele lat,lon,ele
                    </coordinates>
                </LineString>
            </Placemark>
            ...

        </Folder>
	</Document>
</kml>
*/

func PointToCoodinates(point Point) string {
	return fmt.Sprintf("%v,%v,%v", point.Lon, point.Lat, point.Ele)
}
func PointsToCoodinates(points []Point) string {
	w := strings.Builder{}
	for i, point := range points {
		if i > 0 {
			w.WriteString(" ")
		}
		w.WriteString(fmt.Sprintf("%v,%v,%v", point.Lon, point.Lat, point.Ele))
	}
	return w.String()
}

var kmlColors = []struct{ Name, Color string }{
	{"red", "961400FF"},
	{"green", "9678FF00"},
	{"blue", "96FF7800"},
	{"cyan", "96F0FF14"},
	{"orange", "961478FF"},
	{"dark_green", "96008C14"},
	{"purple", "96FF7878"},
	{"pink", "96A078F0"},
	{"brown", "96143C96"},
	{"dark_blue", "96F01414"},
}

func GpxToKml(g gpx) kml {

	var styles []*KmlStyle
	for _, c := range kmlColors {
		styles = append(styles, &KmlStyle{
			Id: c.Name,
			LineStyle: KmlLineStyle{
				Color: c.Color,
				Width: 4,
			},
		})
	}

	var folders []*KmlFolder
	if len(g.Waypoints) > 0 {
		waypointFolder := &KmlFolder{
			Name:        "Waypoints",
			Description: "",
			Visibility:  1,
			Open:        0,
		}
		for _, w := range g.Waypoints {
			waypointFolder.Placemarks = append(waypointFolder.Placemarks, &KmlPlacemark{
				Name:        w.Name,
				Description: w.Desc,
				Visibility:  1,
				Open:        0,

				Point: &KmlPoint{
					Coordinates: PointToCoodinates(w.Point),
				},
			})
		}
		folders = append(folders, waypointFolder)
	}
	if len(g.Routes) > 0 {
		routesFolder := &KmlFolder{
			Name:        "Routes",
			Description: "",
			Visibility:  1,
			Open:        0,
		}
		//for i, r := range g.Routes {
		for _, r := range g.Routes {
			routesFolder.Placemarks = append(routesFolder.Placemarks, &KmlPlacemark{
				Name:        r.Name,
				Description: r.Desc,
				Visibility:  0,
				Open:        0,
				//StyleUrl:    fmt.Sprintf("#%s", kmlColors[i%len(kmlColors)].Name),
				//StyleUrl: "#blue",
				LineString: &KmlLineString{
					Extrude:      true,
					Tessellate:   true,
					AltitudeMode: "clampToGround",
					Coordinates:  PointsToCoodinates(r.Points),
				},
				Style: &KmlStyle{
					LineStyle: KmlLineStyle{
						Color: "#9678FF00",
						//Width: 2,
					},
				},
			})
		}
		folders = append(folders, routesFolder)
	}

	k := kml{
		Xmlns: "http://www.opengis.net/kml/2.2",
		Document: KmlDocument{
			Name:        "Great Himalaya Trail",
			Description: "",
			Visibility:  1,
			Open:        1,
			Styles:      styles,
			Folders:     folders,
		},
	}
	return k
}

type kml struct {
	Xmlns    string      `xml:"xmlns,attr"`
	Document KmlDocument `xml:"Document"`
}

type KmlDocument struct {
	Name        string       `xml:"name"`
	Description string       `xml:"description"`
	Visibility  int          `xml:"visibility"`
	Open        int          `xml:"open"`
	Styles      []*KmlStyle  `xml:"Style"`
	Folders     []*KmlFolder `xml:"Folder"`
}

type KmlStyle struct {
	Id        string       `xml:"id,attr,omitempty"`
	LineStyle KmlLineStyle `xml:"LineStyle"`
}

type KmlLineStyle struct {
	Color string `xml:"color"`
	Width int    `xml:"width,omitempty"`
}

type KmlFolder struct {
	Name        string          `xml:"name"`
	Description string          `xml:"description"`
	Visibility  int             `xml:"visibility"`
	Open        int             `xml:"open"`
	Placemarks  []*KmlPlacemark `xml:"Placemark"`
}

type KmlPlacemark struct {
	Name        string         `xml:"name"`
	Description string         `xml:"description"`
	Visibility  int            `xml:"visibility"`
	Open        int            `xml:"open"`
	StyleUrl    string         `xml:"styleUrl,omitempty"`
	Point       *KmlPoint      `xml:"Point,omitempty"`
	LineString  *KmlLineString `xml:"LineString,omitempty"`
	Style       *KmlStyle      `xml:"Style"`
}

type KmlPoint struct {
	Coordinates string `xml:"coordinates"`
}

type KmlLineString struct {
	Extrude      bool   `xml:"extrude"`
	Tessellate   bool   `xml:"tessellate"`
	AltitudeMode string `xml:"altitudeMode"`
	Coordinates  string `xml:"coordinates"`
}
