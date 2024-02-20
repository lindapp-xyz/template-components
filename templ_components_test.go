package templcomponents_test

import (
	"errors"
	"html/template"
	"io"
	"testing"

	"github.com/lindapp-xyz/template-components"
)

type tcase struct {
	description string
	components  []tcomponent
	input       string
	expected    string
}

type tcomponent struct {
	name     string
	template string
}

var tests = []tcase{
	{
		description: "simple self enclosed component",
		components: []tcomponent{
			{
				name:     "test-component",
				template: `<div id="test-component"><span>a test</span></div>`,
			},
		},
		input:    `<test-component />`,
		expected: `<div id="test-component"><span>a test</span></div>`,
	},
	{
		description: "simple self enclosed component with attributes",
		components: []tcomponent{
			{
				name:     "test-component",
				template: `<div id="test-component"><span>{{.message}}</span></div>`,
			},
		},
		input:    `<test-component message="a test" />`,
		expected: `<div id="test-component"><span>a test</span></div>`,
	},
	{
		description: "simple self enclosed component with attributes using binding",
		components: []tcomponent{
			{
				name:     "test-component",
				template: `<div id="test-component"><span>{{.message}}</span></div>`,
			},
		},
		input:    `<test-component message="{{.message}}" />`,
		expected: `<div id="test-component"><span>{{.message}}</span></div>`,
	},
	{
		description: "components with children",
		components: []tcomponent{
			{
				name:     "test-component",
				template: `<div id="test-component">{{.children}}</div>`,
			},
		},
		input:    `<test-component><span>hello</span></test-component>`,
		expected: `<div id="test-component"><span>hello</span></div>`,
	},
	{
		description: "nested components",
		components: []tcomponent{
			{
				name:     "test-component",
				template: `<div id="test-component">{{.children}}</div>`,
			},
			{
				name:     "nested-component",
				template: `<div id="nested-component">{{.children}}</div>`,
			},
		},
		input:    `<test-component><nested-component><span>hello</span></nested-component></test-component>`,
		expected: `<div id="test-component"><div id="nested-component"><span>hello</span></div></div>`,
	},
	{
		description: "spreading attributes",
		components: []tcomponent{
			{
				name:     "test-component",
				template: `<div id="test-component" {{range $k, $v := .attributes}}{{$k}}="{{$v}}" {{end}}>{{.children}}</div>`,
			},
		},
		input:    `<test-component class="test" data-cy="test"><span>hello</span></test-component>`,
		expected: `<div id="test-component" class="test" data-cy="test" ><span>hello</span></div>`,
	},
	{
		description: "open close tag without children",
		components: []tcomponent{
			{
				name:     "test-component",
				template: `<div id="test-component">{{.children}}</div>`,
			},
		},
		input:    `<test-component></test-component>`,
		expected: `<div id="test-component"></div>`,
	},
}

func TestThing(t *testing.T) {
	for _, test := range tests {
		reg := templcomponents.New()

		for _, comp := range test.components {
            c := templcomponents.Component{
                Name: comp.name,
                Template: template.Must(template.New(comp.name).Parse(comp.template)),
            }
			reg.Add(c)
		}

		res, err := reg.Convert(test.input)

		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatal(err)
		}

		if res != test.expected {
			t.Fatalf("[%s]: Expected \"%s\", got \"%s\"", test.description, test.expected, res)
		}
	}
}
