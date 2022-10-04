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
            <tr><td port="push" sides="ltr"><b>Factory</b></td></tr>
            <tr><td port="switch" align="left">
                Name string<br/>
            </td></tr>
            <tr><td port="switch" align="left">
            </td></tr>
        </table>>
        shape=plain
    ]

    "Mechanic" [
        fillcolor="#88ff0022"
        label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
            <tr><td port="push" sides="ltr"><b>Mechanic</b></td></tr>
            <tr><td port="switch" align="left">
                Skills []string<br/>
                Colleagues []*Mechanic<br/>
            </td></tr>
            <tr><td port="switch" align="left">
            </td></tr>
        </table>>
        shape=plain
    ]

    "Manager" [
        fillcolor="#88ff0022"
        label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
            <tr><td port="push" sides="ltr"><b>Manager</b></td></tr>
            <tr><td port="switch" align="left">
                Pointer *Mechanic<br/>
            </td></tr>
            <tr><td port="switch" align="left">
            </td></tr>
        </table>>
        shape=plain
    ]

    "tool" [
        fillcolor="#88ff0022"
        label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
            <tr><td port="push" sides="ltr"><b>tool</b></td></tr>
            <tr><td port="switch" align="left">
                name string<br/>
            </td></tr>
            <tr><td port="switch" align="left">
            </td></tr>
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
	actual, err := internal.LoadPackages("../examples", "", "")
	for _, p := range actual {
		err := internal.ParsePackage(p, "", "")
		assertEqual(t, nil, err)
	}
	expected := map[string]*internal.Directory{
		"../examples/cars": {
			Path: "../examples/cars",
			Packages: map[string]*internal.Package{
				"cars": {
					Name:       "cars",
					ModulePath: "../examples/cars",
					Files: []*internal.File{
						{
							Path: "../examples/cars/car.go",
							Imports: []*internal.Import{
								{Name: "", Path: "sync", StdLib: true},
								{Name: "", Path: "github.com/slavsan/gog/examples/other", StdLib: false},
							},
							Structs: []*internal.Struct{
								{
									Name: "Camaro",
									Fields: []*internal.Field{
										{Name: "", Type: "other.Vehicle"},
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
				"main": {
					Name:       "main",
					ModulePath: "../examples/cars",
					Files: []*internal.File{
						{
							Path:             "../examples/cars/main.go",
							Imports:          []*internal.Import{},
							BuildConstraints: []string{"mytag"},
							Structs: []*internal.Struct{
								{
									Name: "Foo",
									Fields: []*internal.Field{
										{Name: "Bar", Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		},
		"../examples/other": {
			Path: "../examples/other",
			Packages: map[string]*internal.Package{
				"other": {
					Name:       "other",
					ModulePath: "../examples/other",
					Files: []*internal.File{
						{
							Path:    "../examples/other/vehicle.go",
							Imports: []*internal.Import{},
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
			},
		},
		"../examples": {
			Path: "../examples",
			Packages: map[string]*internal.Package{
				"auto": {
					Name:       "auto",
					ModulePath: "../examples",
					Files: []*internal.File{
						{
							Path: "../examples/factory.go",
							Imports: []*internal.Import{
								{Name: "carmodel", Path: "github.com/slavsan/gog/examples/cars", StdLib: false},
							},
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
			},
		},
	}
	assertEqual(t, nil, err)
	assertEqual(t, expected, actual)
}

func TestFormatPackages(t *testing.T) {
	actual, err := internal.LoadPackages("../examples", "", "")
	assertEqual(t, nil, err)
	for _, p := range actual {
		err := internal.ParsePackage(p, "", "")
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

    subgraph cluster____examples {
        label = "../examples"

        "Factory" [
            fillcolor="#88ff0022"
            label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                <tr><td port="push" sides="ltr"><b>Factory</b></td></tr>
                <tr><td port="switch" align="left">
                    Name string<br/>
                </td></tr>
                <tr><td port="switch" align="left">
                </td></tr>
            </table>>
            shape=plain
        ]

        "Mechanic" [
            fillcolor="#88ff0022"
            label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                <tr><td port="push" sides="ltr"><b>Mechanic</b></td></tr>
                <tr><td port="switch" align="left">
                    Skills []string<br/>
                    Colleagues []*Mechanic<br/>
                </td></tr>
                <tr><td port="switch" align="left">
                </td></tr>
            </table>>
            shape=plain
        ]

        "Manager" [
            fillcolor="#88ff0022"
            label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                <tr><td port="push" sides="ltr"><b>Manager</b></td></tr>
                <tr><td port="switch" align="left">
                    Pointer *Mechanic<br/>
                </td></tr>
                <tr><td port="switch" align="left">
                </td></tr>
            </table>>
            shape=plain
        ]

        "tool" [
            fillcolor="#88ff0022"
            label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                <tr><td port="push" sides="ltr"><b>tool</b></td></tr>
                <tr><td port="switch" align="left">
                    name string<br/>
                </td></tr>
                <tr><td port="switch" align="left">
                </td></tr>
            </table>>
            shape=plain
        ]
    }

    subgraph cluster____examples_cars {
        label = "../examples/cars"

        "Camaro" [
            fillcolor="#88ff0022"
            label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                <tr><td port="push" sides="ltr"><b>Camaro</b></td></tr>
                <tr><td port="switch" align="left">
                    other.Vehicle<br/>
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
                </td></tr>
                <tr><td port="switch" align="left">
                </td></tr>
            </table>>
            shape=plain
        ]
    }

    subgraph cluster____examples_cars {
        label = "../examples/cars"

        "Foo" [
            fillcolor="#88ff0022"
            label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                <tr><td port="push" sides="ltr"><b>Foo</b></td></tr>
                <tr><td port="switch" align="left">
                    Bar string<br/>
                </td></tr>
                <tr><td port="switch" align="left">
                </td></tr>
            </table>>
            shape=plain
        ]
    }

    subgraph cluster____examples_other {
        label = "../examples/other"

        "Vehicle" [
            fillcolor="#88ff0022"
            label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
                <tr><td port="push" sides="ltr"><b>Vehicle</b></td></tr>
                <tr><td port="switch" align="left">
                    Doors int<br/>
                </td></tr>
                <tr><td port="switch" align="left">
                    StartEngine() error<br/>
                    StopEngine() error<br/>
                </td></tr>
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

func TestFormatImports(t *testing.T) {
	actual, err := internal.LoadPackages("../examples", "", "")
	assertEqual(t, nil, err)
	for _, p := range actual {
		err := internal.ParsePackage(p, "", "")
		assertEqual(t, nil, err)
	}
	expected := `digraph {
    rankdir="LR"

    "../examples" -> "github.com/slavsan/gog/examples/cars"
    "../examples/cars" -> "github.com/slavsan/gog/examples/other"
}
`

	actualLines := strings.Split(internal.FormatImports(actual), "\n")
	expectedLines := strings.Split(expected, "\n")

	assertEqual(t, len(expectedLines), len(actualLines))
	for i := range expectedLines {
		assertEqual(t, expectedLines[i], strings.ReplaceAll(actualLines[i], "\t", "        "), fmt.Sprintf("failed on line %d", i))
	}
}

func TestFormatImportsTable(t *testing.T) {
	actual, err := internal.LoadPackages("../examples", "", "")
	assertEqual(t, nil, err)
	for _, p := range actual {
		err := internal.ParsePackage(p, "", "")
		assertEqual(t, nil, err)
	}
	expected := "" +
		fmt.Sprintf("1 %sgithub.com/slavsan/gog/examples/cars%s\n", internal.Green, internal.NoColor) +
		fmt.Sprintf("1 %sgithub.com/slavsan/gog/examples/other%s\n", internal.Green, internal.NoColor) +
		fmt.Sprintf("1 %ssync%s\n", internal.Yellow, internal.NoColor)

	actualLines := strings.Split(internal.FormatImportsTable(actual, "github.com/slavsan/gog"), "\n")
	expectedLines := strings.Split(expected, "\n")

	assertEqual(t, len(expectedLines), len(actualLines))
	for i := range expectedLines {
		assertEqual(t, expectedLines[i], strings.ReplaceAll(actualLines[i], "\t", "        "), fmt.Sprintf("failed on line %d", i))
	}
}

func TestFormatTypes(t *testing.T) {
	actual, err := internal.LoadPackages("../examples", "", "")
	assertEqual(t, nil, err)
	for _, p := range actual {
		err := internal.ParsePackage(p, "", "")
		assertEqual(t, nil, err)
	}
	expected := "" +
		"__YELLOW__../examples__NOCOLOR__\n" +
		"\n" +
		"__GREEN__+__NOCOLOR__ type __BLUE__Factory__NOCOLOR__ {\n" +
		"    __GREEN__+__NOCOLOR__ Name string\n" +
		"}\n" +
		"\n" +
		"__GREEN__+__NOCOLOR__ type __BLUE__Manager__NOCOLOR__ {\n" +
		"    __GREEN__+__NOCOLOR__ Pointer *Mechanic\n" +
		"}\n" +
		"\n" +
		"__GREEN__+__NOCOLOR__ type __BLUE__Mechanic__NOCOLOR__ {\n" +
		"    __GREEN__+__NOCOLOR__ Skills []string\n" +
		"    __GREEN__+__NOCOLOR__ Colleagues []*Mechanic\n" +
		"}\n" +
		"\n" +
		"__RED__-__NOCOLOR__ type __BLUE__tool__NOCOLOR__ {\n" +
		"    __RED__-__NOCOLOR__ name string\n" +
		"}\n" +
		"\n" +
		"__YELLOW__../examples/cars__NOCOLOR__\n" +
		"\n" +
		"__GREEN__+__NOCOLOR__ type __BLUE__Camaro__NOCOLOR__ {\n" +
		"    __RED__-__NOCOLOR__ other.Vehicle\n" +
		"    __RED__-__NOCOLOR__ sync.Mutex\n" +
		"    __GREEN__+__NOCOLOR__ Struct struct{ XXX int }\n" +
		"    __GREEN__+__NOCOLOR__ One string\n" +
		"    __GREEN__+__NOCOLOR__ Fuel interface{}\n" +
		"    __GREEN__+__NOCOLOR__ ChNoPos chan string\n" +
		"    __GREEN__+__NOCOLOR__ ChRecv <-chan int32\n" +
		"    __GREEN__+__NOCOLOR__ ChSend chan<- int32\n" +
		"    __GREEN__+__NOCOLOR__ Features map[string]int\n" +
		"    __GREEN__+__NOCOLOR__ Callback func(string, int) (int64, error)\n" +
		"    __GREEN__+__NOCOLOR__ Two string\n" +
		"    __GREEN__+__NOCOLOR__ Ellipsis func(...string)\n" +
		"    __GREEN__+__NOCOLOR__ ExampleMutex func(sync.Mutex)\n" +
		"    __GREEN__+__NOCOLOR__ Three sync.Mutex\n" +
		"    __GREEN__+__NOCOLOR__ Four sync.Mutex\n" +
		"    __GREEN__+__NOCOLOR__ AnotherStruct struct{  sync.Mutex }\n" +
		"    __GREEN__+__NOCOLOR__ Name string\n" +
		"}\n" +
		"\n" +
		"__YELLOW__../examples/cars__NOCOLOR__\n" +
		"\n" +
		"__GREEN__+__NOCOLOR__ type __BLUE__Foo__NOCOLOR__ { __RED__mytag__NOCOLOR__\n" +
		"    __GREEN__+__NOCOLOR__ Bar string\n" +
		"}\n" +
		"\n" +
		"__YELLOW__../examples/other__NOCOLOR__\n" +
		"\n" +
		"__GREEN__+__NOCOLOR__ type __BLUE__Vehicle__NOCOLOR__ {\n" +
		"    __GREEN__+__NOCOLOR__ Doors int\n" +
		"\n" +
		"    __GREEN__+__NOCOLOR__ StartEngine() error\n" +
		"    __GREEN__+__NOCOLOR__ StopEngine() error\n" +
		"}\n" +
		"\n"

	expected = strings.ReplaceAll(expected, "__GREEN__", internal.Green)
	expected = strings.ReplaceAll(expected, "__YELLOW__", internal.Yellow)
	expected = strings.ReplaceAll(expected, "__RED__", internal.Red)
	expected = strings.ReplaceAll(expected, "__BLUE__", internal.Blue)
	expected = strings.ReplaceAll(expected, "__NOCOLOR__", internal.NoColor)

	actualLines := strings.Split(internal.FormatTypes(actual, "github.com/slavsan/gog"), "\n")
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
		detailed(t, expected, actual, "")
	}
}

func detailed(t *testing.T, expected, actual any, path string) {
	t.Helper()
	var value1 reflect.Value
	var value2 reflect.Value
	if v, ok := expected.(reflect.Value); ok {
		value1 = v
	} else {
		value1 = reflect.ValueOf(expected)
	}
	if v, ok := actual.(reflect.Value); ok {
		value2 = v
	} else {
		value2 = reflect.ValueOf(actual)
	}
	kind1 := value1.Kind()
	kind2 := value2.Kind()
	if kind1 != kind2 {
		t.Errorf(
			"diff: %s (%#v) - %s (%#v) (%s)",
			kind1, expected, kind2, actual, path,
		)
	}
	switch kind1 {
	case reflect.Pointer:
		detailed(t, reflect.Indirect(value1), reflect.Indirect(value2), fmt.Sprintf("%s*", path))
	case reflect.Struct:
		if value1.NumField() != value2.NumField() {
			t.Errorf("struct fields have different length: %s", path)
		}
		num := value1.NumField()
		for i := 0; i < num; i++ {
			detailed(t, value1.Field(i), value2.Field(i), fmt.Sprintf("%s%s.%s", path, value1.Type().String(), value1.Type().Field(i).Name))
		}
	case reflect.Slice:
		if value1.IsNil() != value2.IsNil() {
			t.Errorf(
				"one slice is nil whilst the other is not: %s\n\texpected: %#v\n%s\n\t  actual: %#v\n%s\n",
				path, value1, expand(value1), value2, expand(value2),
			)
		}
		if value1.Len() != value2.Len() {
			t.Errorf(
				"slices have different lengths: %s\n\texpected: %#v\n%s\n\t  actual: %#v\n%s\n",
				path, value1, expand(value1), value2, expand(value2),
			)
		}
		for i := 0; i < value1.Len(); i++ {
			detailed(t, value1.Index(i), value2.Index(i), fmt.Sprintf("%s[%d]", path, i))
		}
	case reflect.Map:
		keys1 := value1.MapKeys()
		keys2 := value2.MapKeys()
		if len(keys1) != len(keys2) {
			t.Errorf("map keys have different lenghts: %v != %v (%s)", keys1, keys2, path)
			return
		}
		for _, v := range value1.MapKeys() {
			detailed(t, value1.MapIndex(v), value2.MapIndex(v), fmt.Sprintf("%s.map[\"%s\"]", path, v))
		}
	case reflect.String:
		if value1.String() != value2.String() {
			t.Errorf(
				"strings are not the same: %s\n\texpected: %s\n\t  actual: %s\n",
				path, value1.String(), value2.String(),
			)
		}
	case reflect.Bool:
		if value1.Bool() != value2.Bool() {
			t.Errorf("bools are not the same: %s", path)
		}
	case reflect.Int:
		if value1.Int() != value2.Int() {
			t.Errorf("ints are not the same: %v != %v %s", value1.Int(), value2.Int(), path)
		}
	default:
		fmt.Printf("UNKNOWN KIND: %s\n", kind1)
	}
}

func expand(v reflect.Value) string {
	var sb strings.Builder
	if v.Type().Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			sb.WriteString(fmt.Sprintf("\t            %#v\n", v.Index(i)))
		}
		return sb.String()
	}
	return sb.String()
}
