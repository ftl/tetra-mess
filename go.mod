module github.com/ftl/tetra-mess

go 1.24.2

//replace github.com/spf13/cobra => ../cobra

//replace github.com/ftl/tetra-pei => ../tetra-pei

require (
	github.com/ftl/tetra-pei v1.2.1
	github.com/hedhyw/Go-Serial-Detector v1.0.0-rc1
	github.com/im7mortal/UTM v1.4.0
	github.com/jacobsa/go-serial v0.0.0-20180131005756-15cf729a72d4
	github.com/spf13/cobra v1.9.1
	github.com/tkrajina/gpxgo v1.4.0
	github.com/twpayne/go-kml/v3 v3.3.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)
