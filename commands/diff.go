package commands

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/aryann/difflib"
	"github.com/concourse/atc"
	"github.com/mgutz/ansi"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v2"
)

type Index interface {
	FindEquivalent(interface{}) (interface{}, bool)
	Slice() []interface{}
}

type Diffs []Diff

type Diff struct {
	Before interface{}
	After  interface{}
}

func name(v interface{}) string {
	return reflect.ValueOf(v).FieldByName("Name").String()
}

func (diff Diff) WriteTo(to io.Writer, label string) {
	indent := gexec.NewPrefixedWriter("  ", to)

	if diff.Before != nil && diff.After != nil {
		fmt.Fprintf(to, ansi.Color("%s %s has changed:", "yellow")+"\n", label, name(diff.Before))

		payloadA, _ := yaml.Marshal(diff.Before)
		payloadB, _ := yaml.Marshal(diff.After)

		renderDiff(indent, string(payloadA), string(payloadB))
	} else if diff.Before != nil {
		fmt.Fprintf(to, ansi.Color("%s %s has been removed:", "yellow")+"\n", label, name(diff.Before))

		payloadA, _ := yaml.Marshal(diff.Before)

		renderDiff(indent, string(payloadA), "")
	} else {
		fmt.Fprintf(to, ansi.Color("%s %s has been added:", "yellow")+"\n", label, name(diff.After))

		payloadB, _ := yaml.Marshal(diff.After)

		renderDiff(indent, "", string(payloadB))
	}
}

type GroupIndex atc.GroupConfigs

func (index GroupIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index GroupIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.GroupConfigs(index).Lookup(name(obj))
}

type JobIndex atc.JobConfigs

func (index JobIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index JobIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.JobConfigs(index).Lookup(name(obj))
}

type ResourceIndex atc.ResourceConfigs

func (index ResourceIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index ResourceIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.ResourceConfigs(index).Lookup(name(obj))
}

type PluginIndex atc.PluginConfigs

func (index PluginIndex) Slice() []interface{} {
	slice := make([]interface{}, len(index))
	for i, object := range index {
		slice[i] = object
	}

	return slice
}

func (index PluginIndex) FindEquivalent(obj interface{}) (interface{}, bool) {
	return atc.PluginConfigs(index).Lookup(name(obj))
}

func diffIndices(oldIndex Index, newIndex Index) Diffs {
	diffs := Diffs{}

	for _, thing := range oldIndex.Slice() {
		newThing, found := newIndex.FindEquivalent(thing)
		if !found {
			diffs = append(diffs, Diff{
				Before: thing,
				After:  nil,
			})
			continue
		}

		if !reflect.DeepEqual(thing, newThing) {
			diffs = append(diffs, Diff{
				Before: thing,
				After:  newThing,
			})
		}
	}

	for _, thing := range newIndex.Slice() {
		_, found := oldIndex.FindEquivalent(thing)
		if !found {
			diffs = append(diffs, Diff{
				Before: nil,
				After:  thing,
			})
			continue
		}
	}

	return diffs
}

func renderDiff(to io.Writer, a, b string) {
	diffs := difflib.Diff(strings.Split(a, "\n"), strings.Split(b, "\n"))

	for _, diff := range diffs {
		text := diff.Payload

		switch diff.Delta {
		case difflib.RightOnly:
			fmt.Fprintf(to, "%s\n", ansi.Color(text, "green"))
		case difflib.LeftOnly:
			fmt.Fprintf(to, "%s\n", ansi.Color(text, "red"))
		case difflib.Common:
			fmt.Fprintf(to, "%s\n", text)
		}
	}
}
