package templcomponents

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Registry struct {
	components map[string]Component
}

func New() *Registry {
	return &Registry{components: make(map[string]Component)}
}

func (r *Registry) Add(c Component) {
	r.components[c.Name] = c
}

type componentTag struct {
	tag     string
	attrs   map[string]interface{}
	content bytes.Buffer
}

// ErrTemplateExecution is returned when a template fails to execute
var ErrTemplateExecution = errors.New("error executing template")

func (r *Registry) Convert(content string) (string, error) {
	var buf bytes.Buffer
	t := html.NewTokenizer(strings.NewReader(content))

	var componentsToClose []componentTag
	currentBuffer := &buf

	for {
		tt := t.Next()
		switch tt {
		case html.ErrorToken:
			err := t.Err()
			if errors.Is(err, io.EOF) {
				return buf.String(), nil
			}
			return buf.String(), err
		case html.SelfClosingTagToken:
			tagName, hasAttrs := t.TagName()
			if comp, ok := r.components[string(tagName)]; ok {
				templData := make(map[string]interface{})
				if hasAttrs {
					for {
						key, val, more := t.TagAttr()
						templData[string(key)] = template.HTMLAttr(val)
						if !more {
							break
						}
					}
				}
				attributes := make(map[template.HTMLAttr]interface{})
				for k, v := range templData {
					attributes[template.HTMLAttr(k)] = v
				}
				templData["attributes"] = attributes
				err := comp.Template.Execute(currentBuffer, templData)
				if err != nil {
					return buf.String(), errors.Join(ErrTemplateExecution, fmt.Errorf("tagname: %s", tagName), err)
				}
			} else {
				_, err := currentBuffer.Write(t.Raw())
				if err != nil {
					return buf.String(), err
				}
			}
		case html.StartTagToken:
			tagName, hasAttrs := t.TagName()
			if _, ok := r.components[string(tagName)]; ok {
				templData := make(map[string]interface{})
				if hasAttrs {
					for {
						key, val, more := t.TagAttr()
						templData[string(key)] = template.HTMLAttr(val)
						if !more {
							break
						}
					}
				}
				attributes := make(map[template.HTMLAttr]interface{})
				for k, v := range templData {
					attributes[template.HTMLAttr(k)] = v
				}
				templData["attributes"] = attributes
				componentsToClose = append(componentsToClose, componentTag{tag: string(tagName), attrs: templData})
			} else {
				_, err := currentBuffer.Write(t.Raw())
				if err != nil {
					return buf.String(), err
				}
			}
		case html.EndTagToken:
			tagName, _ := t.TagName()
			_, isComponent := r.components[string(tagName)]
			if len(componentsToClose) > 0 && isComponent {
				comp := componentsToClose[len(componentsToClose)-1]
				if comp.tag != string(tagName) {
					return buf.String(), errors.New("mismatched end tag")
				}
				componentsToClose = componentsToClose[:len(componentsToClose)-1]

				comp.attrs["children"] = template.HTML(comp.content.String())
				if len(componentsToClose) > 0 {
					currentBuffer = &componentsToClose[len(componentsToClose)-1].content
				} else {
					currentBuffer = &buf
				}
				err := r.components[comp.tag].Template.Execute(currentBuffer, comp.attrs)
				if err != nil {
					return buf.String(), errors.Join(ErrTemplateExecution, fmt.Errorf("tagname: %s", tagName), err)
				}
			} else {
				_, err := currentBuffer.Write(t.Raw())
				if err != nil {
					return buf.String(), err
				}
			}

		default:
			_, err := currentBuffer.Write(t.Raw())
			if err != nil {
				return buf.String(), err
			}
		}

		if len(componentsToClose) > 0 {
			currentBuffer = &componentsToClose[len(componentsToClose)-1].content
		} else {
			currentBuffer = &buf
		}
	}
}

type Component struct {
	Name     string
	Template *template.Template
}
