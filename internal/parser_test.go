package internal_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/slavsan/gog/internal"
)

func TestLoadStructs(t *testing.T) {
	actual, err := internal.LoadStructs("../examples/factory.go")
	expected := []*internal.Struct{
		{
			Name: "Factory",
			Fields: []*internal.Field{
				{Name: "Name", Type: "string"},
			},
		},
		{
			Name: "Mechanic",
			Fields: []*internal.Field{
				{Name: "Skills", Type: "[]string"},
				{Name: "Colleagues", Type: "[]*Mechanic"},
			},
		},
		{
			Name: "Manager",
			Fields: []*internal.Field{
				{Name: "Pointer", Type: "*Mechanic"},
			},
		},
		{
			Name: "tool",
			Fields: []*internal.Field{
				{Name: "name", Type: "string"},
			},
		},
	}
	assertEqual(t, nil, err)
	assertEqual(t, expected, actual)
}

func TestFormatStructs(t *testing.T) {
	actual, err := internal.LoadStructs("../examples/factory.go")
	assertEqual(t, nil, err)

	expected := `digraph {
        graph [
                label = "Basic git concepts and operations\n\n"
                labelloc = t
                fontname = "Helvetica,Arial,sans-serif"
                fontsize = 20
                layout = dot
                rankdir = LR
                newrank = true
        ]

        node [
                style=filled
                shape=rect
                pencolor="#00000044" // frames color
                fontname="Helvetica,Arial,sans-serif"
                shape=plaintext
        ]


        "Factory" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Factory</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Name string<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]

        "Mechanic" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Mechanic</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Skills []string<br/>
Colleagues []*Mechanic<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]

        "Manager" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Manager</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Pointer *Mechanic<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]

        "tool" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>tool</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                name string<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]
}
`
	actualLines := strings.Split(internal.Format(actual), "\n")
	expectedLines := strings.Split(expected, "\n")

	assertEqual(t, len(expectedLines), len(actualLines))
	for i := range expectedLines {
		assertEqual(t, expectedLines[i], strings.ReplaceAll(actualLines[i], "\t", "        "), fmt.Sprintf("failed on line %d", i))
	}
}

func TestLoadPackages(t *testing.T) {
	actual, err := internal.LoadPackages("../examples")
	for _, p := range actual {
		err := internal.ParsePackage(p)
		assertEqual(t, nil, err)
	}
	expected := map[string]*internal.Package{
		"../examples": {
			Name: "auto",
			Path: "../examples",
			Files: []*internal.File{
				{
					Path: "../examples/factory.go",
					Structs: []*internal.Struct{
						{
							Name: "Factory",
							Fields: []*internal.Field{
								{Name: "Name", Type: "string"},
							},
						},
						{
							Name: "Mechanic",
							Fields: []*internal.Field{
								{Name: "Skills", Type: "[]string"},
								{Name: "Colleagues", Type: "[]*Mechanic"},
							},
						},
						{
							Name: "Manager",
							Fields: []*internal.Field{
								{Name: "Pointer", Type: "*Mechanic"},
							},
						},
						{
							Name: "tool",
							Fields: []*internal.Field{
								{Name: "name", Type: "string"},
							},
						},
					},
				},
			},
		},
		"../examples/cars": {
			Name: "cars",
			Path: "../examples/cars",
			Files: []*internal.File{
				{
					Path: "../examples/cars/car.go",
					Structs: []*internal.Struct{
						{
							Name: "Camaro",
							Fields: []*internal.Field{
								{Name: "Name", Type: "string"},
								{Name: "Features", Type: "map[string]int"},
								{Name: "Callback", Type: "func(string, int) (int64, error)"},
								{Name: "Fuel", Type: "interface{}"},
								{Name: "ChNoPos", Type: "chan string"},
								{Name: "ChRecv", Type: "<-chan int32"},
								{Name: "ChSend", Type: "chan<- int32"},
								{Name: "Struct", Type: "struct{ XXX int }"},
								{Name: "One", Type: "string"},
								{Name: "Two", Type: "string"},
								{Name: "Ellipsis", Type: "func(...string)"},
								{Name: "ExampleMutex", Type: "func(sync.Mutex)"},
								{Name: "Three", Type: "sync.Mutex"},
								{Name: "Four", Type: "sync.Mutex"},
								{Name: "AnotherStruct", Type: "struct{  sync.Mutex }"},
								{Name: "", Type: "sync.Mutex"},
							},
						},
					},
				},
			},
		},
		"../examples/other": {
			Name: "other",
			Path: "../examples/other",
			Files: []*internal.File{
				{
					Path: "../examples/other/vehicle.go",
					Structs: []*internal.Struct{
						{
							Name: "Vehicle",
							Fields: []*internal.Field{
								{Name: "Doors", Type: "int"},
							},
							Methods: []*internal.Method{
								{Signature: "StartEngine() error"},
								{Signature: "StopEngine() error"},
							},
						},
					},
				},
			},
		},
	}
	assertEqual(t, nil, err)
	assertEqual(t, expected, actual)
}

func TestFormatPackages(t *testing.T) {
	actual, err := internal.LoadPackages("../examples")
	assertEqual(t, nil, err)
	for _, p := range actual {
		err := internal.ParsePackage(p)
		assertEqual(t, nil, err)
	}
	expected := `digraph {
        graph [
                label = "Basic git concepts and operations\n\n"
                labelloc = t
                fontname = "Helvetica,Arial,sans-serif"
                fontsize = 20
                layout = dot
                rankdir = LR
                newrank = true
        ]

        node [
                style=filled
                shape=rect
                pencolor="#00000044" // frames color
                fontname="Helvetica,Arial,sans-serif"
                shape=plaintext
        ]
subgraph cluster____examples {label = "../examples"


        "Factory" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Factory</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Name string<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]

        "Mechanic" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Mechanic</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Skills []string<br/>
Colleagues []*Mechanic<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]

        "Manager" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Manager</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Pointer *Mechanic<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]

        "tool" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>tool</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                name string<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]
}

subgraph cluster____examples_cars {label = "../examples/cars"


        "Camaro" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Camaro</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Name string<br/>
Features map[string]int<br/>
Callback func(string, int) (int64, error)<br/>
Fuel interface{}<br/>
ChNoPos chan string<br/>
ChRecv &lt;-chan int32<br/>
ChSend chan&lt;- int32<br/>
Struct struct{ XXX int }<br/>
One string<br/>
Two string<br/>
Ellipsis func(...string)<br/>
ExampleMutex func(sync.Mutex)<br/>
Three sync.Mutex<br/>
Four sync.Mutex<br/>
AnotherStruct struct{  sync.Mutex }<br/>
 sync.Mutex<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                
                        </td> </tr>
                </table>>
                shape=plain
        ]
}

subgraph cluster____examples_other {label = "../examples/other"


        "Vehicle" [
                fillcolor="#88ff0022"
                label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                        <tr> <td port="push" sides="ltr"><b>Vehicle</b></td> </tr>
                        <tr> <td port="switch" align="left">
                                Doors int<br/>

                        </td> </tr>
                        <tr> <td port="switch" align="left">
                                StartEngine() error<br/>
StopEngine() error<br/>

                        </td> </tr>
                </table>>
                shape=plain
        ]
}
}
`

	actualLines := strings.Split(internal.FormatPackages(actual), "\n")
	expectedLines := strings.Split(expected, "\n")

	assertEqual(t, len(expectedLines), len(actualLines))
	for i := range expectedLines {
		assertEqual(t, expectedLines[i], strings.ReplaceAll(actualLines[i], "\t", "        "), fmt.Sprintf("failed on line %d", i))
	}
}

func assertEqual(t *testing.T, expected, actual any, msg ...string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		message := ""
		if len(msg) > 0 {
			message = fmt.Sprintf("\t message: %s", strings.Join(msg, "\n"))
		}
		t.Errorf(
			"equality assertion failed:\n\texpected: %#v (%s)\n\t  actual: %#v (%s)\n%s",
			expected, reflect.TypeOf(expected),
			actual, reflect.TypeOf(actual),
			message,
		)
	}
}
