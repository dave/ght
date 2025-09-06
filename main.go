package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {

	// Calculate total stats for all legs (paste into google sheet)
	//if err := CalcStats(); err != nil {
	//	panic(err)
	//}

	// Process final routes and output new file (remember to increment version)
	const VERSION = 9
	if err := ProcessFinalRoutesAll(VERSION); err != nil {
		panic(err)
	}

	if err := CreateTrailNotes(VERSION); err != nil {
		panic(err)
	}

	// Create map images for trail notes
	//if err := DrawMaps(68, 69, 70); err != nil {
	//	panic(err)
	//}

	// Create map elevation graphs for trail notes
	//if err := DrawElevations(68, 69, 70); err != nil {
	//	panic(err)
	//}
}

func ProcessFinalRoutesAll(version int) error {
	if err := ProcessFinalRoutes(true, version); err != nil {
		return err
	}
	if err := ProcessFinalRoutes(false, version); err != nil {
		return err
	}
	return nil
}

func ProcessFinalRoutes(mapsme bool, version int) error {

	b, err := ioutil.ReadFile(`/Users/dave/src/ght/trailnotes.json`)
	if err != nil {
		return err
	}
	var notes TrailNotesSheetStruct
	if err := json.Unmarshal(b, &notes); err != nil {
		return err
	}

	legsByLeg := map[int]*LegStruct{}

	legs := notes.Legs
	for i, leg := range legs {
		if i == 0 {
			leg.From = "Taplejung"
		} else {
			leg.From = legs[i-1].To
		}
		for _, waypoint := range notes.Waypoints {
			if waypoint.Leg == leg.Leg {
				leg.Waypoints = append(leg.Waypoints, waypoint)
			}
		}
		for _, pass := range notes.Passes {
			if pass.Leg == leg.Leg {
				leg.Passes = append(leg.Passes, pass)
			}
		}
		if leg.Vlog != nil {
			days := strings.Split(fmt.Sprint(leg.Vlog), ",")
			for _, day := range days {
				d, err := strconv.Atoi(day)
				if err != nil {
					return err
				}
				leg.Days = append(leg.Days, d)
			}
		}
		legsByLeg[leg.Leg] = leg
	}

	inDir := `/Users/dave/Library/CloudStorage/Dropbox/Adventures/2019 GHT Nepal/GPX files corrected with waypoints`
	outDir := `/Users/dave/Library/CloudStorage/Dropbox/Adventures/2019 GHT Nepal/GPX files corrected final`

	routeFiles, err := ioutil.ReadDir(inDir)
	if err != nil {
		return err
	}

	out := gpx{
		Version: 1.1,
	}

	for _, fileInfo := range routeFiles {
		if !strings.HasSuffix(fileInfo.Name(), ".gpx") {
			continue
		}
		g := loadGpx(filepath.Join(inDir, fileInfo.Name()))
		legNumber, err := strconv.Atoi(fileInfo.Name()[1:4])
		if err != nil {
			return err
		}
		//fmt.Println(legNumber, len(g.Routes[0].Points), len(g.Waypoints))

		leg := legsByLeg[legNumber]

		// check all waypoints
		for _, waypointFromNotes := range leg.Waypoints {
			var found bool
			for _, waypointFromGpx := range g.Waypoints {
				if fmt.Sprintf("L%03d %s", legNumber, waypointFromNotes.Name) == waypointFromGpx.Name {
					found = true
					waypointFromNotes.Lat = waypointFromGpx.Lat
					waypointFromNotes.Lon = waypointFromGpx.Lon
					waypointFromNotes.Elevation = waypointFromGpx.Ele
					//fmt.Printf("%d\t%s\t%f\n", legNumber, waypointFromNotes.Name, waypointFromGpx.Ele)
					break
				}
			}
			if !found {
				return fmt.Errorf("missing waypoint %s", waypointFromNotes.Name)
			}
		}

		for _, waypointFromGpx := range g.Waypoints {
			var found bool
			for _, waypointFromNotes := range leg.Waypoints {
				if fmt.Sprintf("L%03d %s", legNumber, waypointFromNotes.Name) == waypointFromGpx.Name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("missing waypoint %s", waypointFromGpx.Name)
			}
		}

		for _, pass := range leg.Passes {
			var found bool
			for _, waypointFromGpx := range g.Waypoints {
				if fmt.Sprintf("L%03d %s", legNumber, pass.Pass) == waypointFromGpx.Name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("missing pass %s", pass.Pass)
			}
		}
		routeDesc := fmt.Sprintf("%s", leg.Notes)
		if mapsme {
			routeDesc = ""
		}
		var points []Point
		if len(g.Routes) > 0 {
			points = g.Routes[0].Points
		} else {
			// gpx.studio converts all routes to tracks, so we must handle some input files with tracks
			for _, point := range g.Tracks[0].Segments[0].Points {
				points = append(points, Point{point.Lat, point.Lon, point.Ele})
			}
		}
		out.Routes = append(out.Routes, Route{
			Name:   fmt.Sprintf("L%03d %s to %s", leg.Leg, leg.From, leg.To),
			Desc:   routeDesc,
			Points: points,
		})
		if mapsme {
			// maps.me doesn't show descriptions for routes so we add a dummy waypoint and remove the route desc

			var lat, lon, ele float64
			if len(g.Routes) > 0 {
				lat = g.Routes[0].Points[0].Lat
				lon = g.Routes[0].Points[0].Lon
				ele = g.Routes[0].Points[0].Ele
			} else {
				lat = g.Tracks[0].Segments[0].Points[0].Lat
				lon = g.Tracks[0].Segments[0].Points[0].Lon
				ele = g.Tracks[0].Segments[0].Points[0].Ele
			}
			out.Waypoints = append(out.Waypoints, Waypoint{
				Point: Point{
					Lat: lat,
					Lon: lon,
					Ele: ele,
				},
				Name: fmt.Sprintf("L%03d %s to %s", leg.Leg, leg.From, leg.To),
				Desc: leg.Notes,
			})
		}
		for _, w := range leg.Waypoints {
			out.Waypoints = append(out.Waypoints, Waypoint{
				Point: Point{
					Lat: w.Lat,
					Lon: w.Lon,
					Ele: w.Elevation,
				},
				Name: fmt.Sprintf("L%03d %s", leg.Leg, w.Name),
				Desc: w.Notes,
			})
		}

	}

	if mapsme {
		saveKml(GpxToKml(out), filepath.Join(outDir, fmt.Sprintf("routes-for-maps-me-v%v.kml", version)))
	} else {
		saveGpx(out, filepath.Join(outDir, fmt.Sprintf("routes-v%v.gpx", version)))
		saveKml(GpxToKml(out), filepath.Join(outDir, fmt.Sprintf("routes-v%v.kml", version)))
	}

	return nil

}

func CalcStats() error {

	routesDir := "/Users/dave/Library/CloudStorage/Dropbox/Adventures/2019 GHT Nepal/GPX files corrected with waypoints"
	routeFiles, err := ioutil.ReadDir(routesDir)
	if err != nil {
		return err
	}

	for _, fileInfo := range routeFiles {

		if !strings.HasSuffix(fileInfo.Name(), ".gpx") {
			continue
		}

		leg, err := strconv.Atoi(fileInfo.Name()[1:4])
		if err != nil {
			return err
		}

		if leg != 72 {
			//continue
		}

		g := loadGpx(filepath.Join(routesDir, fileInfo.Name()))
		if len(g.Routes)+len(g.Tracks) != 1 {
			return fmt.Errorf("not 1 route / track for %q", fileInfo.Name())
		}

		var points []Point
		if len(g.Routes) > 0 {
			points = g.Routes[0].Points
		} else {
			// gpx.studio converts all routes to tracks, so we must handle some input files with tracks
			for _, point := range g.Tracks[0].Segments[0].Points {
				points = append(points, Point{point.Lat, point.Lon, point.Ele})
			}
		}

		var length, climb, descent, start, end, top, bottom float64
		for i, current := range points {
			if i == 0 {
				start = current.Ele
				top = current.Ele
				bottom = current.Ele
			}
			if i == len(points)-1 {
				end = current.Ele
			}
			if i == 0 {
				continue
			}

			// work out distance delta
			previous := points[i-1]
			horizontal := distance(current.Lat, current.Lon, previous.Lat, previous.Lon)

			// work out elevation delta
			var vertical float64
			climbing := current.Ele > previous.Ele
			if climbing {
				vertical = current.Ele - previous.Ele
			} else {
				vertical = previous.Ele - current.Ele
			}

			if vertical > 50 {
				// discard outlier points
				continue
			}

			verticalkm := vertical / 1000.0

			total := math.Sqrt(horizontal*horizontal + verticalkm*verticalkm)

			if current.Ele > top {
				top = current.Ele
			}
			if current.Ele < bottom {
				bottom = current.Ele
			}

			length += total

			if climbing {
				climb += vertical
			} else {
				descent += vertical
			}
		}
		//fmt.Printf("%d\t%f\t%f\t%f\t%f\t%f\t%f\t%f\n", leg, length, climb, descent, start, end, top, bottom)
		fmt.Printf("%f\t%f\t%f\t%f\t%f\t%f\t%f\n", length, climb, descent, start, end, top, bottom)
		//start = end
		//end = start
		//if climb > 3000 {
		//fmt.Printf("%d\t%f\t%f\n", leg, length, climb)
		//}
	}

	return nil
}

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515

	dist = dist * 1.609344

	return dist
}

func saveGpx(g gpx, filename string) {
	bw, err := xml.MarshalIndent(g, "", "\t")
	//bw, err := xml.Marshal(g)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filename, []byte(xml.Header+string(bw)), 0777); err != nil {
		panic(err)
	}
}

func saveKml(k kml, filename string) {
	bw, err := xml.MarshalIndent(k, "", "\t")
	//bw, err := xml.Marshal(k)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filename, []byte(xml.Header+string(bw)), 0777); err != nil {
		panic(err)
	}
}

func loadGpx(filename string) gpx {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("error reading file %q: %w", filename, err))
	}
	var g gpx
	if err := xml.NewDecoder(bytes.NewBuffer(b)).Decode(&g); err != nil {
		panic(fmt.Errorf("error decoding xml for %q: %w", filename, err))
	}
	return g
}

func closest(points []Point, p Point) int {
	minDist := -1.0
	minIndex := 0
	for k, v := range points {
		d := distance(v.Lat, v.Lon, p.Lat, p.Lon)
		if d < minDist || minDist == -1.0 {
			minDist = d
			minIndex = k
		}
	}
	return minIndex
}

type gpx struct {
	Version   float64    `xml:"version,attr"`
	Waypoints []Waypoint `xml:"wpt"`
	Tracks    []Track    `xml:"trk"`
	Routes    []Route    `xml:"rte"`
}

type Waypoint struct {
	Point
	Name string `xml:"name"`
	Sym  string `xml:"sym,omitempty"`
	Desc string `xml:"desc"`
}

type Route struct {
	Name   string  `xml:"name"`
	Desc   string  `xml:"desc"`
	Points []Point `xml:"rtept"`
}

type Point struct {
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
	Ele float64 `xml:"ele"`
}

type TrackPoint struct {
	Point
	Time *time.Time `xml:"time,omitempty"`
}

type Track struct {
	Name     string         `xml:"name"`
	Desc     string         `xml:"desc"`
	Segments []TrackSegment `xml:"trkseg"`
}

type TrackSegment struct {
	Points []TrackPoint `xml:"trkpt"`
}

type TrailNotesSheetStruct struct {
	Legs      []*LegStruct
	Waypoints []*WaypointStruct
	Passes    []*PassStruct
}

type LegStruct struct {
	Leg  int
	Vlog interface{}

	To                                              string
	Length, Climb, Descent, Start, End, Top, Bottom float64
	Route, Trail, Quality                           int
	Lodge                                           string
	Notes                                           string

	From      string
	Waypoints []*WaypointStruct
	Passes    []*PassStruct
	Days      []int

	RouteString, TrailString, LodgeString, QualityString string
}

type WaypointStruct struct {
	Leg                 int
	Name, Notes         string
	Lat, Lon, Elevation float64
}

type PassStruct struct {
	Leg    int
	Pass   string
	Height float64
}
